from typing import List, Optional, Dict, Any, Union
from pydantic import BaseModel, Field

# Plugin-specific configuration models
class CodeDumpPluginConfig(BaseModel):
    project_path: Optional[str] = Field(
        None,
        description="Path to the project directory to scan"
    )
    use_gitignore: bool = Field(
        True,
        description="Whether to respect .gitignore patterns"
    )
    max_depth: int = Field(
        3,
        description="Maximum directory depth to include"
    )
    include_content: bool = Field(
        False,
        description="Whether to include file contents or just structure"
    )
    file_types: Optional[List[str]] = Field(
        None,
        description="List of file extensions to include (e.g., ['.py', '.md'])"
    )

class DataServicePluginConfig(BaseModel):
    base_path: Optional[str] = Field(
        None,
        description="Path to the data directory"
    )
    include_all_files: bool = Field(
        True,
        description="Whether to include all files regardless of type"
    )
    file_types: Optional[List[str]] = Field(
        None,
        description="List of file extensions to include if not including all (e.g., ['.json', '.csv'])"
    )
    header: str = Field(
        "USER DATA FILES",
        description="Text to use as the header for the data section"
    )

# Generic container for plugin configurations
class PluginConfig(BaseModel):
    code_dump: Optional[CodeDumpPluginConfig] = Field(
        None,
        description="Configuration for the code_dump plugin"
    )
    data_service: Optional[DataServicePluginConfig] = Field(
        None,
        description="Configuration for the data_service plugin"
    )

    class Config:
        schema_extra = {
            "example": {
                "code_dump": {
                    "project_path": "/path/to/project",
                    "use_gitignore": True,
                    "include_content": False
                },
                "data_service": {
                    "base_path": "data",
                    "include_all_files": True
                }
            }
        }

# More intuitive Chat Request model
class ChatRequest(BaseModel):
    prompt: str = Field(
        ...,
        description="The prompt to send to the model",
        example="Tell me about my project structure and data"
    )

    # Plugin selection
    plugins: Optional[List[str]] = Field(
        None,
        description="List of plugin names to activate (e.g., ['code_dump', 'data_service'])",
        example=["code_dump", "data_service"]
    )

    # For activating with configurations
    plugin_config: Optional[PluginConfig] = Field(
        None,
        description="Configuration for specific plugins"
    )

    reset_plugins: bool = Field(
        False,
        description="Whether to reset all plugins before activating the specified ones"
    )

# Legacy format support - can be used internally but hidden from docs
class PluginRequest(BaseModel):
    name: str
    config: Optional[Dict[str, Any]] = None