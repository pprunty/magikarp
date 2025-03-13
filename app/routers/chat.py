import logging
from fastapi import Depends, APIRouter, HTTPException, status
from app.services.model import TransformerModel
from app.dependencies import get_transformer_model
from app.models.requests import ChatRequest, PluginRequest
from app.models.responses import ChatResponse, ErrorResponse

# Configure the logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('uvicorn.debug')

chat_router = APIRouter(prefix="/chat", tags=["chat"])

@chat_router.post("", response_model=ChatResponse, status_code=status.HTTP_200_OK,
                  responses={400: {"model": ErrorResponse}, 500: {"model": ErrorResponse}})
async def chat_with_ai_on_device(
        request: ChatRequest,
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Get a response from the conversation model with configurable plugins.

    This endpoint allows you to send a prompt to the model with optional plugin configuration.
    You can specify which plugins to activate, provide custom configuration for each plugin,
    and choose to reset all plugins before applying new ones.

    ## Plugin Configuration

    ### Code Dump Plugin
    Adds project structure to the prompt for code understanding:
    ```json
    "code_dump": {
        "project_path": "/path/to/project",
        "use_gitignore": true,
        "max_depth": 3,
        "include_content": false,
        "file_types": [".py", ".md"]
    }
    ```

    ### Data Service Plugin
    Adds user data files to the prompt:
    ```json
    "data_service": {
        "base_path": "data",
        "include_all_files": true,
        "file_types": [".json", ".csv"],
        "header": "USER DATA FILES"
    }
    ```

    ## Examples

    ### Basic prompt without plugins:
    ```json
    {
        "prompt": "What is the capital of France?"
    }
    ```

    ### Using code_dump plugin:
    ```json
    {
        "prompt": "Explain my project architecture",
        "plugins": ["code_dump"],
        "reset_plugins": true
    }
    ```

    ### Using data_service plugin with configuration:
    ```json
    {
        "prompt": "Analyze my data",
        "plugin_config": {
            "data_service": {
                "include_all_files": false,
                "file_types": [".json"]
            }
        },
        "reset_plugins": true
    }
    ```

    Args:
        request: The chat request with optional plugin configuration.
        transformer_model: The transformer model dependency.

    Returns:
        A ChatResponse object containing the model's response.

    Raises:
        HTTPException: If the prompt is empty or an error occurs.
    """
    logger.info(f"Received chat request: {request}")

    if not request.prompt:
        logger.error("Prompt cannot be empty")
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Prompt cannot be empty.")

    try:
        # Handle plugin activation/deactivation
        if request.reset_plugins:
            # Get currently active plugins
            active_plugins = transformer_model.plugin_manager.active_plugins.copy()

            # Deactivate all plugins
            for plugin in active_plugins:
                transformer_model.deactivate_plugin(plugin)
                logger.info(f"Deactivated plugin: {plugin}")

        # Configure and activate plugins from plugin_config
        if request.plugin_config:
            # Process code_dump plugin if configured
            if request.plugin_config.code_dump:
                config_dict = request.plugin_config.code_dump.dict(exclude_none=True)
                if config_dict:
                    success = transformer_model.register_plugin("code_dump", **config_dict)
                    if not success:
                        logger.warning("Failed to register code_dump plugin")

                # Activate the plugin
                success = transformer_model.activate_plugin("code_dump")
                if success:
                    logger.info("Activated code_dump plugin")
                else:
                    logger.warning("Failed to activate code_dump plugin")

            # Process data_service plugin if configured
            if request.plugin_config.data_service:
                config_dict = request.plugin_config.data_service.dict(exclude_none=True)
                if config_dict:
                    success = transformer_model.register_plugin("data_service", **config_dict)
                    if not success:
                        logger.warning("Failed to register data_service plugin")

                # Activate the plugin
                success = transformer_model.activate_plugin("data_service")
                if success:
                    logger.info("Activated data_service plugin")
                else:
                    logger.warning("Failed to activate data_service plugin")

        # Activate specified plugins (without custom config)
        elif request.plugins:
            for plugin_name in request.plugins:
                success = transformer_model.activate_plugin(plugin_name)
                if success:
                    logger.info(f"Activated plugin: {plugin_name}")
                else:
                    logger.warning(f"Failed to activate plugin: {plugin_name}")

        # Send the prompt to the model
        response = transformer_model.ask_model(request.prompt)
        logger.info(f"Generated response: {response}")

        return {"response": response}
    except Exception as e:
        logger.error(f"Error generating response: {e}")
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=f"Error generating response: {str(e)}")