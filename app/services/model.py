import json
import logging
import ollama
from enum import Enum
from typing import List
from app.enums.rules import RuleSetEnum
from app.plugins.manager import PluginManager
from datetime import date

# Configure the logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('uvicorn.debug')

class TransformerModel:
    """Service class for leveraging Ollama's use of the Llama3 transformer model."""

    def __init__(self, plugin_manager: PluginManager = None):
        """
        Initializes the TransformerModel with plugin manager.

        Args:
            plugin_manager (PluginManager, optional): Plugin manager instance for prompt processing
        """
        self.chat_messages = []
        self.dynamic_rules = list(RuleSetEnum)
        self.plugin_manager = plugin_manager or PluginManager()

    def ask_model(self, prompt: str) -> str:
        """
        Asks the model a question based on the given prompt.
        Processes the prompt through active plugins before sending to the model.

        Args:
            prompt (str): The input prompt for the model.

        Returns:
            str: The model's response.
        """
        # Process the prompt through all active plugins
        processed_prompt = self.plugin_manager.process_prompt(prompt)

        user_message = {'role': 'user', 'content': processed_prompt}
        self.chat_messages.append(user_message)
        logger.info(f"Chat messages: {self.chat_messages}")

        try:
            response = ollama.chat(model='magikarp', messages=self.chat_messages)
            return response['message']['content']
        except Exception as e:
            logger.error(f"Error in asking model: {e}")
            raise

    def generate_suggested_prompts(self) -> dict:
        """
        Generates predefined prompt suggestions.

        Returns:
            dict: A dictionary of suggested prompts.

        Raises:
            ValueError: If the response cannot be parsed as JSON.
        """
        prompt = (
            "Can you provide me with predefined prompt suggestions to ask an AI assistant? "
            "Please give your response in JSON format where the keys in the response are numbers and values the suggested "
            "prompt."
        )

        # Process the prompt through all active plugins
        processed_prompt = self.plugin_manager.process_prompt(prompt)

        user_message = {'role': 'user', 'content': processed_prompt}
        self.chat_messages.append(user_message)
        logger.info(f"Chat messages: {self.chat_messages}")

        try:
            response = ollama.chat(model='magikarp', messages=self.chat_messages)
            response_content = response['message']['content']
            logger.info(f"Suggested prompts response: {response_content}")

            # Extract the JSON part from the response
            start_index = response_content.find('{')
            end_index = response_content.rfind('}') + 1
            json_str = response_content[start_index:end_index]
            return json.loads(json_str)
        except (json.JSONDecodeError, KeyError, IndexError) as e:
            logger.error(f"Error generating suggested prompts: {e}")
            raise ValueError("Failed to parse suggested prompts from the response")

    def get_dynamic_rules(self) -> List[str]:
        """
        Returns the list of all dynamic rules.

        Returns:
            List[str]: The descriptions of all dynamic rules.
        """
        return [rule.value for rule in self.dynamic_rules]

    def register_plugin(self, plugin_name: str, **kwargs) -> bool:
        """
        Register a plugin with the transformer model.

        Args:
            plugin_name (str): Name of the plugin to register
            **kwargs: Parameters to pass to the plugin constructor

        Returns:
            bool: True if registration successful, False otherwise
        """
        return self.plugin_manager.register_plugin(plugin_name, **kwargs)

    def activate_plugin(self, plugin_name: str) -> bool:
        """
        Activate a registered plugin.

        Args:
            plugin_name (str): Name of the plugin to activate

        Returns:
            bool: True if activation successful, False otherwise
        """
        return self.plugin_manager.activate_plugin(plugin_name)

    def deactivate_plugin(self, plugin_name: str) -> bool:
        """
        Deactivate an active plugin.

        Args:
            plugin_name (str): Name of the plugin to deactivate

        Returns:
            bool: True if deactivation successful, False otherwise
        """
        return self.plugin_manager.deactivate_plugin(plugin_name)