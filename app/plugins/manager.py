from typing import Dict, List, Type, Optional
import logging
from app.plugins.base import PluginBase
from app.plugins import PLUGINS

logger = logging.getLogger(__name__)

class PluginManager:
    """Manager for handling plugins that modify prompts."""

    def __init__(self):
        """Initialize the PluginManager."""
        self.plugins: Dict[str, PluginBase] = {}
        self.active_plugins: List[str] = []

    def register_plugin(self, plugin_name: str, **kwargs) -> bool:
        """
        Register a plugin with the manager.

        Args:
            plugin_name (str): Name of the plugin to register
            **kwargs: Parameters to pass to the plugin constructor

        Returns:
            bool: True if registration successful, False otherwise
        """
        if plugin_name not in PLUGINS:
            logger.error(f"Plugin '{plugin_name}' not found")
            return False

        try:
            plugin_class = PLUGINS[plugin_name]
            plugin_instance = plugin_class(**kwargs)
            self.plugins[plugin_name] = plugin_instance
            logger.info(f"Plugin '{plugin_name}' registered successfully")
            return True
        except Exception as e:
            logger.error(f"Error registering plugin '{plugin_name}': {e}")
            return False

    def activate_plugin(self, plugin_name: str) -> bool:
        """
        Activate a registered plugin.

        Args:
            plugin_name (str): Name of the plugin to activate

        Returns:
            bool: True if activation successful, False otherwise
        """
        if plugin_name not in self.plugins:
            logger.error(f"Cannot activate '{plugin_name}': Plugin not registered")
            return False

        if plugin_name not in self.active_plugins:
            self.active_plugins.append(plugin_name)
            logger.info(f"Plugin '{plugin_name}' activated")
        return True

    def deactivate_plugin(self, plugin_name: str) -> bool:
        """
        Deactivate an active plugin.

        Args:
            plugin_name (str): Name of the plugin to deactivate

        Returns:
            bool: True if deactivation successful, False otherwise
        """
        if plugin_name in self.active_plugins:
            self.active_plugins.remove(plugin_name)
            logger.info(f"Plugin '{plugin_name}' deactivated")
            return True
        return False

    def process_prompt(self, prompt: str) -> str:
        """
        Process a prompt through all active plugins.

        Args:
            prompt (str): The original prompt

        Returns:
            str: The modified prompt after processing by all active plugins
        """
        modified_prompt = prompt

        for plugin_name in self.active_plugins:
            if plugin_name in self.plugins:
                plugin = self.plugins[plugin_name]
                try:
                    modified_prompt = plugin.process_prompt(modified_prompt)
                    logger.debug(f"Prompt processed by plugin '{plugin_name}'")
                except Exception as e:
                    logger.error(f"Error processing prompt with plugin '{plugin_name}': {e}")

        return modified_prompt