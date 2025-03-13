from typing import List, Optional, Dict, Any
from pydantic import BaseModel

class PluginListResponse(BaseModel):
    available_plugins: List[str]
    active_plugins: List[str]

class CodeDumpConfig(BaseModel):
    project_path: Optional[str] = None
    use_gitignore: bool = True
    max_depth: int = 3
    include_content: bool = False
    file_types: Optional[List[str]] = None

class DataServiceConfig(BaseModel):
    base_path: Optional[str] = None
    include_all_files: bool = True
    file_types: Optional[List[str]] = None
    header: str = "USER DATA FILES"

# Define a mapping of plugin names to their config models
# This allows us to validate plugin-specific configurations
plugin_config_models = {
    "code_dump": CodeDumpConfig,
    "data_service": DataServiceConfig
}

# Generic plugin configuration model that accepts any plugin type
class PluginConfig(BaseModel):
    plugin_name: str
    config: Dict[str, Any]