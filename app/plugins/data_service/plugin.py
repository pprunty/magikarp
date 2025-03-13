"""Plugin for providing user data files to the model."""
import os
import csv
import json
import logging
from typing import Dict, List, Any, Optional
from app.plugins.base import PluginBase

# Configure the logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('uvicorn.debug')

class DataServicePlugin(PluginBase):
    """Plugin that adds user data files to prompts."""

    def __init__(self, base_path: str = 'data', include_all_files: bool = True,
                 file_types: Optional[List[str]] = None, header: str = "USER DATA FILES"):
        """
        Initialize the DataServicePlugin.

        Args:
            base_path (str): The base directory where data files are located
            include_all_files (bool): Whether to include all files or only specific types
            file_types (list, optional): List of file extensions to include (e.g., ['.json', '.csv'])
            header (str): Header text to use for the data section
        """
        self.base_path = os.path.join(os.getcwd(), base_path)
        self.include_all_files = include_all_files
        self.file_types = file_types
        self.header = header
        self._file_contents = {}
        self._validate_data_directory()

    @property
    def name(self) -> str:
        return "data_service"

    def _validate_data_directory(self):
        """Validates that the data directory exists."""
        if not os.path.exists(self.base_path):
            logger.error(f"Data directory {self.base_path} does not exist.")
            raise FileNotFoundError(f"Data directory {self.base_path} does not exist.")
        logger.info(f"Data directory {self.base_path} found.")

    def _read_file_as_text(self, file_path: str) -> str:
        """Reads a file and returns its contents as text, handling different file types appropriately.

        Args:
            file_path (str): Path to the file to read.

        Returns:
            str: The contents of the file as text.
        """
        file_extension = os.path.splitext(file_path)[1].lower()

        try:
            # JSON files
            if file_extension == '.json':
                with open(file_path, 'r') as file:
                    data = json.load(file)
                    return json.dumps(data, indent=2)

            # CSV files
            elif file_extension == '.csv':
                rows = []
                with open(file_path, 'r') as file:
                    csv_reader = csv.reader(file)
                    for row in csv_reader:
                        rows.append(row)

                if not rows:
                    return ""

                # Format CSV as text table
                result = []
                header = rows[0]
                for row in rows:
                    formatted_row = []
                    for i, cell in enumerate(row):
                        if i < len(header):
                            formatted_row.append(f"{header[i]}: {cell}")
                    result.append(", ".join(formatted_row))
                return "\n".join(result)

            # All other files (text, etc.)
            else:
                with open(file_path, 'r') as file:
                    return file.read()

        except Exception as e:
            logger.error(f"Error reading file {file_path}: {e}")
            return f"[Error reading file: {str(e)}]"

    def load_all_files(self) -> Dict[str, str]:
        """Loads all files from the data directory.

        Returns:
            Dict[str, str]: A dictionary mapping file names to their contents.
        """
        self._file_contents = {}

        for file_name in os.listdir(self.base_path):
            file_path = os.path.join(self.base_path, file_name)

            # Skip directories and non-readable files
            if os.path.isdir(file_path):
                logger.info(f"Skipping directory: {file_name}")
                continue

            # Skip files not matching the file_types if specified
            if not self.include_all_files and self.file_types:
                file_extension = os.path.splitext(file_name)[1].lower()
                if file_extension not in self.file_types:
                    logger.info(f"Skipping file with non-matching extension: {file_name}")
                    continue

            try:
                logger.info(f"Loading file: {file_name}")
                self._file_contents[file_name] = self._read_file_as_text(file_path)
            except Exception as e:
                logger.error(f"Could not load file {file_name}: {e}")
                self._file_contents[file_name] = f"[Error loading file: {str(e)}]"

        return self._file_contents

    def get_formatted_data(self) -> str:
        """Formats and returns all data files for output.

        Returns:
            str: The formatted data.
        """
        logger.info("Formatting data for output")

        # Make sure files are loaded
        if not self._file_contents:
            self.load_all_files()

        formatted_data = f"{self.header}\n"
        formatted_data += "=" * len(self.header) + "\n\n"

        for file_name, content in self._file_contents.items():
            formatted_data += f"[FILE: {file_name}]\n"
            formatted_data += "-" * (len(file_name) + 8) + "\n"
            formatted_data += content + "\n\n"

        return formatted_data

    def process_prompt(self, prompt: str) -> str:
        """
        Add user data files to the prompt.

        Args:
            prompt (str): The original prompt

        Returns:
            str: Prompt with user data files
        """
        # Reload files every time to capture any changes
        self.load_all_files()

        # Get formatted data
        formatted_data = self.get_formatted_data()

        # Add the formatted data before the prompt
        return formatted_data + prompt