from app.plugins.base import PluginBase
from app.plugins.code_dump.plugin import CodeDumpPlugin
from app.plugins.data_service.plugin import DataServicePlugin

# Plugin registry to be populated with available plugins
PLUGINS = {
    "code_dump": CodeDumpPlugin,
    "data_service": DataServicePlugin
}