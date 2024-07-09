import json
import ollama
from magikarp.services.data import DataService
from magikarp.static.prompts import PromptsEnum


class TransformerModel:
    def __init__(self):
        self.chat_messages = []
        self.user_data_service = DataService(user="dan")  # < This would be an id in real-life scenario
        self.formatted_user_data = self.user_data_service.get_formatted_data()

    def ask_model(self, prompt):
        self.chat_messages.append({'role': 'user', 'content': self.formatted_user_data + prompt})
        print(self.chat_messages)
        response = ollama.chat(model='magikarp', messages=self.chat_messages)
        return response['message']['content']

    def generate_suggested_prompts(self):
        prompt = (
            "Can you based on my user data, provide me with predfined prompt suggestions to ask an A.I assistant. "
            "Please give your response in JSON format with no additional text and the prompt key number and value the suggested prompt.")
        self.chat_messages.append({'role': 'user', 'content': self.formatted_user_data + prompt})
        response = ollama.chat(model='magikarp', messages=self.chat_messages)
        response_content = response['message']['content']
        print(f"suggested_prompts = {response_content}")
        # Extract the JSON part from the response
        start_index = response_content.find('{')
        end_index = response_content.rfind('}') + 1

        # Convert the JSON string to a dictionary
        json_str = response_content[start_index:end_index]
        return json.loads(json_str)
