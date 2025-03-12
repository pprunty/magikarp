"""
This module defines an Enum for various prompt types and a Pydantic model
that dynamically uses these Enum values to create a Literal type for validation.
"""

from enum import Enum


class PromptsEnum(str, Enum):
    """Enumeration of possible prompts for recommendations."""

    default = "What should I do today?"
    prompt_1 = "What's the best way to stay motivated for my daily runs?"
    prompt_2 = "Can you recommend some new songs to add to my Morning Motivation playlist?"
    prompt_3 = "How can I optimize my Strava profile to get more kudos?"
    prompt_4 = "Do you have any tips on how to balance work and personal life based on my calendar events?"
    prompt_5 = "What's the best way to use my screen time effectively, considering my app usage habits?"
    # Add predefined additional prompts here ...
