from app.plugins.base import PluginBase
from app.plugins.code_dump.plugin import CodeDumpPlugin

# Plugin registry to be populated with available plugins
PLUGINS = {
    "code_dump": CodeDumpPlugin
}
