import logging
from fastapi import APIRouter, HTTPException, Depends, status, Body
from typing import List, Dict, Any, Optional
from app.services.model import TransformerModel
from app.dependencies import get_transformer_model
from app.models.responses import PluginsListResponse, PluginStatusResponse, ErrorResponse
from app.models.requests import CodeDumpPluginConfig, DataServicePluginConfig, PluginRequest

# Configure the logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('uvicorn.debug')

# Create router
router = APIRouter(
    prefix="/plugins",
    tags=["plugins"],
    responses={404: {"model": ErrorResponse}},
)

@router.get("", response_model=PluginsListResponse, status_code=status.HTTP_200_OK,
            summary="List available plugins")
async def get_available_plugins(
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Get a list of available and active plugins.

    Returns:
        A PluginsListResponse object with available and active plugins.
    """
    available_plugins = list(transformer_model.plugin_manager.plugins.keys())
    active_plugins = transformer_model.plugin_manager.active_plugins

    return {
        "available_plugins": available_plugins,
        "active_plugins": active_plugins
    }

@router.get("/{plugin_name}", response_model=PluginStatusResponse,
            status_code=status.HTTP_200_OK,
            responses={404: {"model": ErrorResponse}},
            summary="Get plugin status")
async def get_plugin_status(
        plugin_name: str,
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Get status and configuration of a specific plugin.

    Args:
        plugin_name: Name of the plugin to get status for
        transformer_model: The transformer model dependency

    Returns:
        A PluginStatusResponse object with plugin status and configuration

    Raises:
        HTTPException: If the plugin is not found
    """
    if plugin_name not in transformer_model.plugin_manager.plugins:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Plugin '{plugin_name}' not found"
        )

    plugin_instance = transformer_model.plugin_manager.plugins.get(plugin_name)
    is_active = plugin_name in transformer_model.plugin_manager.active_plugins

    # Get plugin configuration - this is a simplified version
    # In a real implementation, you'd want to extract the actual config from the plugin
    config = {
        "name": plugin_name,
        # Add other config properties here based on plugin type
    }

    return {
        "name": plugin_name,
        "active": is_active,
        "config": config
    }

@router.post("/code_dump/register", response_model=PluginStatusResponse,
             status_code=status.HTTP_200_OK,
             responses={400: {"model": ErrorResponse}},
             summary="Register code_dump plugin with configuration")
async def register_code_dump_plugin(
        config: CodeDumpPluginConfig = Body(...,
                                            example={
                                                "project_path": "/path/to/project",
                                                "use_gitignore": True,
                                                "max_depth": 3,
                                                "include_content": False
                                            }),
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Register the code_dump plugin with custom configuration.

    This plugin adds project structure to prompts, allowing the model to understand your codebase.

    Args:
        config: Configuration for the code_dump plugin
        transformer_model: The transformer model dependency

    Returns:
        A PluginStatusResponse object with plugin status

    Raises:
        HTTPException: If the plugin registration fails
    """
    plugin_name = "code_dump"
    config_dict = config.dict(exclude_none=True)

    success = transformer_model.register_plugin(plugin_name, **config_dict)

    if not success:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Failed to register plugin '{plugin_name}'"
        )

    is_active = plugin_name in transformer_model.plugin_manager.active_plugins

    return {
        "name": plugin_name,
        "active": is_active,
        "config": {
            "name": plugin_name,
            **config_dict
        }
    }

@router.post("/data_service/register", response_model=PluginStatusResponse,
             status_code=status.HTTP_200_OK,
             responses={400: {"model": ErrorResponse}},
             summary="Register data_service plugin with configuration")
async def register_data_service_plugin(
        config: DataServicePluginConfig = Body(...,
                                               example={
                                                   "base_path": "data",
                                                   "include_all_files": True,
                                                   "file_types": [".json", ".csv"],
                                                   "header": "USER DATA FILES"
                                               }),
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Register the data_service plugin with custom configuration.

    This plugin adds user data files to prompts, giving the model context about your data.

    Args:
        config: Configuration for the data_service plugin
        transformer_model: The transformer model dependency

    Returns:
        A PluginStatusResponse object with plugin status

    Raises:
        HTTPException: If the plugin registration fails
    """
    plugin_name = "data_service"
    config_dict = config.dict(exclude_none=True)

    success = transformer_model.register_plugin(plugin_name, **config_dict)

    if not success:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Failed to register plugin '{plugin_name}'"
        )

    is_active = plugin_name in transformer_model.plugin_manager.active_plugins

    return {
        "name": plugin_name,
        "active": is_active,
        "config": {
            "name": plugin_name,
            **config_dict
        }
    }

@router.post("/{plugin_name}/activate", response_model=PluginStatusResponse,
             status_code=status.HTTP_200_OK,
             responses={404: {"model": ErrorResponse}},
             summary="Activate a plugin")
async def activate_plugin(
        plugin_name: str,
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Activate a specific plugin.

    Args:
        plugin_name: Name of the plugin to activate
        transformer_model: The transformer model dependency

    Returns:
        A PluginStatusResponse object with updated plugin status

    Raises:
        HTTPException: If the plugin is not found or cannot be activated
    """
    if plugin_name not in transformer_model.plugin_manager.plugins:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Plugin '{plugin_name}' not found"
        )

    success = transformer_model.activate_plugin(plugin_name)
    if not success:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Failed to activate plugin '{plugin_name}'"
        )

    plugin_instance = transformer_model.plugin_manager.plugins.get(plugin_name)

    # Get plugin configuration - simplified version
    config = {
        "name": plugin_name,
        # Add other config properties here based on plugin type
    }

    return {
        "name": plugin_name,
        "active": True,
        "config": config
    }

@router.post("/{plugin_name}/deactivate", response_model=PluginStatusResponse,
             status_code=status.HTTP_200_OK,
             responses={404: {"model": ErrorResponse}},
             summary="Deactivate a plugin")
async def deactivate_plugin(
        plugin_name: str,
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Deactivate a specific plugin.

    Args:
        plugin_name: Name of the plugin to deactivate
        transformer_model: The transformer model dependency

    Returns:
        A PluginStatusResponse object with updated plugin status

    Raises:
        HTTPException: If the plugin is not found
    """
    if plugin_name not in transformer_model.plugin_manager.plugins:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Plugin '{plugin_name}' not found"
        )

    success = transformer_model.deactivate_plugin(plugin_name)

    plugin_instance = transformer_model.plugin_manager.plugins.get(plugin_name)

    # Get plugin configuration - simplified version
    config = {
        "name": plugin_name,
        # Add other config properties here based on plugin type
    }

    return {
        "name": plugin_name,
        "active": False,
        "config": config
    }

@router.post("/reset", response_model=PluginsListResponse,
             status_code=status.HTTP_200_OK,
             summary="Reset all plugins")
async def reset_plugins(
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Deactivate all active plugins.

    Args:
        transformer_model: The transformer model dependency

    Returns:
        A PluginsListResponse object with updated active plugins (empty list)
    """
    # Get currently active plugins
    active_plugins = transformer_model.plugin_manager.active_plugins.copy()

    # Deactivate all plugins
    for plugin in active_plugins:
        transformer_model.deactivate_plugin(plugin)
        logger.info(f"Deactivated plugin: {plugin}")

    return {
        "available_plugins": list(transformer_model.plugin_manager.plugins.keys()),
        "active_plugins": []
    }

@router.post("/batch", response_model=PluginsListResponse,
             status_code=status.HTTP_200_OK,
             summary="Configure multiple plugins in a single request")
async def batch_configure_plugins(
        plugin_requests: List[PluginRequest],
        reset_first: bool = False,
        transformer_model: TransformerModel = Depends(get_transformer_model)
):
    """Configure multiple plugins in a single request.

    Args:
        plugin_requests: List of plugins to configure
        reset_first: Whether to reset all plugins before configuring
        transformer_model: The transformer model dependency

    Returns:
        A PluginsListResponse object with updated active plugins
    """
    # Reset all plugins if requested
    if reset_first:
        active_plugins = transformer_model.plugin_manager.active_plugins.copy()
        for plugin in active_plugins:
            transformer_model.deactivate_plugin(plugin)
            logger.info(f"Deactivated plugin: {plugin}")

    # Configure and activate each plugin
    for plugin_req in plugin_requests:
        plugin_name = plugin_req.name
        plugin_config = plugin_req.config or {}

        # Re-register the plugin with new configuration if provided
        if plugin_config:
            success = transformer_model.register_plugin(plugin_name, **plugin_config)
            if not success:
                logger.warning(f"Failed to register plugin: {plugin_name}")

        # Activate the plugin
        success = transformer_model.activate_plugin(plugin_name)
        if success:
            logger.info(f"Activated plugin: {plugin_name}")
        else:
            logger.warning(f"Failed to activate plugin: {plugin_name}")

    return {
        "available_plugins": list(transformer_model.plugin_manager.plugins.keys()),
        "active_plugins": transformer_model.plugin_manager.active_plugins
    }