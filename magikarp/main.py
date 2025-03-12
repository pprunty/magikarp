import asyncio
import json
import logging

from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from magikarp.settings import settings

from magikarp.services.model import Model
from magikarp.services.manager import ActionManager  # Assuming this is defined as shown earlier
from magikarp.services.snapshots import SnapshotManager  # The SnapshotManager class

# Configure logging for the websocket module.
logger = logging.getLogger("websocket")
logger.setLevel(logging.DEBUG)
handler = logging.StreamHandler()
formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
handler.setFormatter(formatter)
logger.addHandler(handler)

# Create FastAPI instance
app = FastAPI(
    title="Magikarp API",
    description="APIs for Magikarp service.",
    version="1.0.0"
)

# CORS Middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Adjust as needed
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Pydantic model for the request body (for the REST endpoint)
class PromptRequest(BaseModel):
    prompt: str

# Root endpoint
@app.get("/")
async def read_root():
    logger.info("Root endpoint accessed.")
    return {
        "message": "Hello! Welcome to the Magikarp API.",
        "instructions": "Add '/docs' to the URL in your browser to view and manually trigger the APIs."
    }

# REST endpoint to prompt the agent (for single-shot requests)
@app.post("/agent/prompt")
async def prompt_agent(request: PromptRequest):
    logger.info("Received REST prompt request.")
    model = Model()
    response = model.chat(request.prompt)
    logger.debug(f"REST prompt response: {response}")
    return {"response": response}

# WebSocket endpoint for continuous action chaining
@app.websocket("/ws/chain")
async def websocket_chain(websocket: WebSocket):
    await websocket.accept()
    logger.info("WebSocket connection accepted.")

    model = Model()  # Instantiate the agent model
    # Initialize the ActionManager with a designated project root (adjust as needed)
    action_manager = ActionManager(project_root="~/GitHub/rust-project")
    # Initialize the SnapshotManager using the same project root.
    snapshot = SnapshotManager(project_root="~/GitHub/rust-project")

    # Start with an initial prompt that includes the project outline and current snapshot.
    conversation_context = (
        f"You are building the following project: {settings.PROJECT_OUTLINE}.\n"
        "Based on the current state of the project, please provide the next action.\n"
        "Here is the current state of the project:\n"
        f"{snapshot.get_snapshot()}\n\n"
        "Available actions:\n"
        "1. run_command: Use this action to run a CLI command (e.g., to install dependencies, build, or run the project). "
        "The JSON response should include a 'command' key in the data payload.\n"
        "2. update_file: Use this action to create or update a file in the project. "
        "The JSON response should include 'file_name', 'file_content', 'file_path', and 'new_file' (a boolean) keys in the data payload.\n"
        "Please return your next action as a valid JSON response following one of the examples above."
    )
    logger.info("Initial conversation context set.")

    try:
        while True:
            logger.info("Prompting agent with current conversation context:\n%s", conversation_context)
            # 1. Prompt the agent using the current conversation context.
            action_response = model.chat(conversation_context)
            logger.debug("Agent responded with action: %s", action_response.action)

            # 2. Use the ActionManager to execute the agent's suggested action.
            result = action_manager.handle_action(action_response)
            logger.debug("Action executed with result: %s", result)

            if isinstance(result, dict) and result.get("returncode") == 127:
                stderr = result.get("stderr", "")
                parts = stderr.split(":")
                missing_command = parts[1].strip() if len(parts) > 1 else "unknown command"
                conversation_context += (
                    f"\n\nNote: The command '{missing_command}' is not available in the current environment. "
                    "Please choose an alternative action (such as updating files, installing the required tool, or checking the command syntax)."
                )



           # 3. Update the conversation context with the executed action's details.
            conversation_context += (
                f"\n\nExecuted action: {action_response.action}\n"
                f"Result: {result}\n"
                "Provide the next action."
            )
            logger.info("Updated conversation context.")

            # 4. Send the result of the executed action back to the client.
            outgoing = json.dumps({
                "action": action_response.action,
                "result": result,
                "next_prompt": conversation_context
            })
            await websocket.send_text(outgoing)
            logger.info("Sent action result to client via WebSocket.")

            # 5. Optionally, wait for a client message to either continue or stop.
            try:
                client_message = await asyncio.wait_for(websocket.receive_text(), timeout=10)
                logger.info("Received message from client: %s", client_message)
                if client_message.strip().lower() in ["stop", "exit"]:
                    await websocket.send_text("Terminating session as requested.")
                    logger.info("Terminating WebSocket session per client request.")
                    break
                # Optionally, the client could provide additional context to append.
                conversation_context += f"\nClient: {client_message}"
                logger.debug("Appended client message to conversation context.")
            except asyncio.TimeoutError:
                logger.debug("No client message received within timeout. Continuing session.")
                # If no message is received within the timeout, continue automatically.
                pass

    except WebSocketDisconnect:
        logger.info("WebSocket disconnected.")
