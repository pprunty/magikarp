from typing import Optional

from pydantic import BaseModel
from magikarp.static.prompts import PromptsEnum
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
