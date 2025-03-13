"""
Dependencies module to avoid circular imports between routers and the main app.
"""
from typing import Dict, Any
from fastapi import Request

# This will store our application-wide dependencies
_app_state: Dict[str, Any] = {}

def set_app_state(key: str, value: Any) -> None:
    """Set a value in the application state."""
    _app_state[key] = value

def get_app_state(key: str) -> Any:
    """Get a value from the application state."""
    return _app_state.get(key)

def get_transformer_model():
    """Get the TransformerModel instance."""
    from app.services.model import TransformerModel
    model = get_app_state("transformer_model")
    if not model:
        raise RuntimeError("TransformerModel not initialized in app state")
    return model

# Alias for backward compatibility if needed
get_transformer_service = get_transformer_model