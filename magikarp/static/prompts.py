"""
This module defines an Enum for various prompt types and a Pydantic model
that dynamically uses these Enum values to create a Literal type for validation.
"""

from enum import Enum


class PromptsEnum(str, Enum):
    """Enumeration of possible prompts for recommendations."""

    default = "What should I do today?"
    # The additional prompts will be appended dynamically
