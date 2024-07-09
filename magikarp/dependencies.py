"""
Module to define dependency functions for the FastAPI application.
"""
from magikarp.services.model import TransformerModel


def get_transformer_service() -> TransformerModel:
    """
    Dependency injection function to get the initialized TransformerService.

    Returns:
        TransformerService: The initialized transformer service.
    """
    return TransformerModel()
