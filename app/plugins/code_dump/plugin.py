import os
from app.plugins.base import PluginBase
from app.plugins.code_dump.main import ProjectScanner

class CodeDumpPlugin(PluginBase):
    """Plugin that adds project structure to prompts."""

    def __init__(self, project_path: str = None, use_gitignore: bool = True,
                 max_depth: int = 3, include_content: bool = False,
                 file_types: list = None):
        """
        Initialize the CodeDumpPlugin.

        Args:
            project_path (str, optional): Path to the project directory
            use_gitignore (bool): Whether to respect .gitignore
            max_depth (int): Maximum directory depth to include
            include_content (bool): Whether to include file contents
            file_types (list, optional): List of file extensions to include
        """
        self.project_path = project_path or os.getcwd()
        self.use_gitignore = use_gitignore
        self.max_depth = max_depth
        self.include_content = include_content
        self.file_types = file_types

    @property
    def name(self) -> str:
        return "code_dump"

    def _get_project_structure(self) -> str:
        """Get the project structure using ProjectScanner."""
        scanner = ProjectScanner(
            self.project_path,
            file_types=self.file_types,
            use_gitignore=self.use_gitignore
        )

        if self.include_content:
            # Get full scan with file contents
            return scanner.scan()
        else:
            # Get just the directory structure
            return scanner.get_directory_structure()

    def process_prompt(self, prompt: str) -> str:
        """
        Add project structure to the prompt.

        Args:
            prompt (str): The original prompt

        Returns:
            str: Prompt with project structure
        """
        project_info = self._get_project_structure()

        # Create a formatted project structure section
        project_section = (
            "### Project Structure ###\n"
            f"{project_info}\n"
            "### End Project Structure ###\n\n"
        )

        # Add the project structure before the prompt
        return project_section + prompt
