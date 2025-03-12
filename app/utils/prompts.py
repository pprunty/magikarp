from enum import Enum
from typing import Tuple, Literal

from app.enums.prompts import PromptsEnum


def get_enum_values(enum_cls: Enum) -> Tuple[str, ...]:
    """
    Extracts the values from an Enum class.

    Args:
        enum_cls (Enum): The Enum class from which to extract values.

    Returns:
        Tuple[str, ...]: A tuple containing all the values of the Enum.
    """
    return tuple(item.value for item in enum_cls)


# Dynamically create the Literal type using the enum values
PromptsLiteral = Literal[get_enum_values(PromptsEnum)]
