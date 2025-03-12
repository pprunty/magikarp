import json
import logging
import ollama
from enum import Enum
from typing import List
from app.enums.rules import RuleSetEnum
from app.services.data import DataService
from datetime import date

# Configure the logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('uvicorn.debug')

class TransformerModel:
    """Service class for leveraging Ollama's use of the Llama3 transformer model."""

    def __init__(self):
        """Initializes the TransformerModel with user data."""
        self.chat_messages = []
        self.user_data_service = DataService()  # The user would be fetched via an ID in real-life scenario
        self.formatted_user_data = self.user_data_service.get_formatted_data()
        self.dynamic_rules = list(RuleSetEnum)

    def ask_model(self, prompt: str) -> str:
        """Asks the model a question based on the given prompt and user data.

        Args:
            prompt (str): The input prompt for the model.

        Returns:
            str: The model's response.
        """
        user_message = {'role': 'user', 'content': self.formatted_user_data + prompt}
        self.chat_messages.append(user_message)
        logger.info(f"Chat messages: {self.chat_messages}")

        try:
            response = ollama.chat(model='magikarp', messages=self.chat_messages)
            return response['message']['content']
        except Exception as e:
            logger.error(f"Error in asking model: {e}")
            raise

    def generate_suggested_prompts(self) -> dict:
        """Generates predefined prompt suggestions based on user data.

        Returns:
            dict: A dictionary of suggested prompts.

        Raises:
            ValueError: If the response cannot be parsed as JSON.
        """
        prompt = (
            "Can you, based on my user data, provide me with predefined prompt suggestions to ask an AI assistant? "
            "Please give your response in JSON format where the keys in the response are numbers and values the suggested"
            "prompt."
        )
        user_message = {'role': 'user', 'content': self.formatted_user_data + prompt}
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

    def get_push_notifications(self, date: date, rules: List[str]) -> str:
        """Asks the model for simulated push notifications for a given date based on user defined
        rule set.

        Args:
            date (str): The date for which to generate notifications.
            rules (List[RuleSetEnum]): The list of rules to apply for generating notifications.

        Returns:
            str: The model's response.
        """
        rules_str = "\n".join([rule for rule in rules])
        prompt = (
            f"For the given date {date} and these user defined rules stored as preferences on the user's device:\n"
            f"{rules_str}\n\n"
            f"Provide simulated push notifications to the user for the date using the user's metadata and rules. "
            f"Your response should be in JSON format where the key is the datetime of the notification and the "
            f"value is no more than two sentences detailing the notification."
            f"Please ensure notifications include at least six notifications and include at least one health and social "
            f"encouragement notifications."
        )
        try:
            return self.ask_model(prompt=prompt)
        except Exception as e:
            logger.error(f"Error in asking model: {e}")
            raise

    def add_rule(self, rule_description: str):
        """Dynamically adds a new rule to the rule set.

        Args:
            rule_description (str): The description of the new rule.

        Returns:
            None
        """
        new_rule_name = f"dynamic_rule_{len(self.dynamic_rules) + 1}"
        # Simulate adding the rule to the Enum by appending it to the dynamic list
        self.dynamic_rules.append(Enum(new_rule_name, rule_description))
        logger.info(f"Added new rule: {new_rule_name} = {rule_description}")

    def get_dynamic_rules(self) -> List[str]:
        """Returns the list of all dynamic rules.

        Returns:
            List[str]: The descriptions of all dynamic rules.
        """
        return [rule.value for rule in self.dynamic_rules]