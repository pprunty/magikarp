import logging
from fastapi import Depends, APIRouter, HTTPException, status
from magikarp.services.model import TransformerModel
from magikarp.dependencies import get_transformer_service
from magikarp.models.responses import ChatResponse
from magikarp.models.requests import ChatRequest

# Configure the logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('uvicorn.debug')

chat_router = APIRouter(prefix="/chat", tags=["chat"])

@chat_router.post("", response_model=ChatResponse, status_code=status.HTTP_200_OK)
async def chat_with_ai_on_device(
        request: ChatRequest,
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    """Endpoint to get a response from the conversation model based on a prompt.

    Args:
        request: The input request containing the prompt.
        transformer_service: The transformer service dependency to interact with the model.

    Returns:
        A dictionary with the model's response.

    Raises:
        HTTPException: If the prompt is empty or an error occurs.
    """
    logger.info(f"Received chat request: {request}")

    if not request.prompt:
        logger.error("Prompt cannot be empty")
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Prompt cannot be empty.")

    try:
        response = transformer_service.ask_model(request.prompt)
        logger.info(f"Generated response: {response}")
        return {"response": response}
    except Exception as e:
        logger.error(f"Error generating response: {e}")
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail="Error generating response")
