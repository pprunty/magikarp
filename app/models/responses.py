from typing import List, Dict, Any
from pydantic import BaseModel

class ChatResponse(BaseModel):
    response: str

class ErrorResponse(BaseModel):
    detail: str

class PluginsListResponse(BaseModel):
    available_plugins: List[str]
    active_plugins: List[str]

class PluginStatusResponse(BaseModel):
    name: str
    active: bool
    config: Dict[str, Any]