from typing import List
from pydantic import BaseModel, Field
from datetime import date

from magikarp.enums.rules import RuleSetEnum
from magikarp.enums.prompts import PromptsEnum
from magikarp.utils.prompts import PromptsLiteral


class PromptRequest(BaseModel):
    """
    Pydantic model that validates the 'prompts' field against the values
    defined in PromptsEnum.

    Attributes:
        prompts (PromptsLiteral): A field that can take only the values defined
            in PromptsEnum.
    """
    prompts: PromptsLiteral = PromptsEnum.default.value


class ChatRequest(BaseModel):
    """Simple pydantic model to store prompt to LLM"""
    prompt: str = Field(..., example="How many steps do I have left to do today to meet my step goal?")


class AddRuleRequest(BaseModel):
    """Model to represent a new rule to be added."""
    rule_description: str


class DateRequest(BaseModel):
    """Model to represent the date for which notifications are requested."""
    request_date: date = Field(..., example="2024-05-15")
    rules: List[RuleSetEnum] = Field(..., example=[rule.value for rule in RuleSetEnum], exclude=True)
