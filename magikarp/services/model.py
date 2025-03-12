from ollama import chat
from typing import Union
import json

from magikarp.models.actions import (
    UpdateFile,
    CommandRequest,
    ActionResponse,
)

class Model:
    def __init__(self, model: str = "magikarp:latest"):
        self.model = model
        self.schema = {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string",
                    "enum": [
                        "update_file",
                        "run_command"
                    ],
                    "description": "Specifies the type of action to be performed."
                },
                "data": {
                    "oneOf": [
                        {
                            "type": "object",
                            "properties": {
                                "file_name": {"type": "string"},
                                "file_content": {"type": "string"},
                                "file_path": {"type": "string"},
                                "new_file": {"type": "boolean"}
                            },
                            "required": ["file_name", "file_content", "file_path", "new_file"]
                        },
                        {
                            "type": "object",
                            "properties": {
                                "command": {"type": "string"}
                            },
                            "required": ["command"]
                        }
                    ]
                }
            },
            "required": ["action", "data"]
        }

    def chat(self, prompt: str) -> Union[
        UpdateFile,
        CommandRequest,
    ]:
        response = chat(
            messages=[
                {
                    'role': 'user',
                    'content': prompt,
                }
            ],
            model=self.model,
            format=self.schema
        )
        print(f"response = {response}")
        return self.clean_response(response)

    def clean_response(self, response_obj: Union[str, dict]) -> ActionResponse:
        """
        Parses a raw response into an ActionResponse object.

        Args:
            response_obj (Union[str, dict]): The raw response from the service.

        Returns:
            ActionResponse: The validated action response.
        """
        if isinstance(response_obj, dict):
            message = response_obj.get("message", {})
            if isinstance(message, dict):
                response_str = message.get("content", "")
            else:
                response_str = str(message)
        else:
            response_str = response_obj.strip()

        response_dict = json.loads(response_str)
        return ActionResponse.parse_obj(response_dict)
