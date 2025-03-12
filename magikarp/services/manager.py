import os
import subprocess
import logging
from pathlib import Path
from magikarp.models.actions import (
    ActionResponse,
    UpdateFile,
    CommandRequest,
    CreateDirectory,
    DeleteFile,
)

# Configure a logger for this module.
logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)  # Adjust log level as needed.
handler = logging.StreamHandler()  # You can configure handlers as needed.
formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
handler.setFormatter(formatter)
logger.addHandler(handler)

class ActionManager:
    def __init__(self, project_root: str):
        """
        Args:
            project_root (str): The root directory of the project that the AI agent is building.
        """
        self.project_root = Path(project_root).expanduser().resolve()
        # Ensure the project root exists.
        self.project_root.mkdir(parents=True, exist_ok=True)
        logger.info(f"Initialized ActionManager with project root: {self.project_root}")

    def handle_action(self, action_response: ActionResponse):
        """
        Process an ActionResponse by calling the appropriate method based on the action type.
        """
        action = action_response.action
        data = action_response.data
        logger.info(f"Handling action: {action} with data: {data}")

        if action == "update_file":
            result = self.update_file(data)
        elif action == "run_command":
            result = self.run_command(data)
        elif action == "create_directory":
            result = self.create_directory(data)
        elif action == "delete_file":
            result = self.delete_file(data)
        else:
            logger.error(f"Unsupported action: {action}")
            raise ValueError(f"Unsupported action: {action}")

        logger.info(f"Action '{action}' executed with result: {result}")
        return result

    def update_file(self, data: UpdateFile):
        """
        Create or update a file.

        This method always opens the file in write mode ("w"), which creates the file
        if it doesn't exist or updates it if it does.

        Note: Ignores the 'file_path' field returned in the response and uses only 'file_name'
        as a relative path from the project root.
        """
        # Use file_name as a relative path from the project root.
        file_path = self.project_root / data.file_name
        # Ensure that the parent directory exists.
        file_path.parent.mkdir(parents=True, exist_ok=True)
        logger.debug(f"Preparing to create or update file at: {file_path}")

        # Determine if the file already exists for logging purposes.
        if file_path.exists():
            operation = "updated"
        else:
            operation = "created"

        try:
            with open(file_path, "w") as f:
                f.write(data.file_content)
            logger.info(f"File {operation}: {file_path}")
            return f"File {operation}: {file_path}"
        except Exception as e:
            logger.error(f"Error writing file {file_path}: {e}")
            raise e

    def run_command(self, data: CommandRequest):
        """
        Run a shell command in the project root.
        """
        logger.debug(f"Running command: {data.command} in {self.project_root}")
        # Use subprocess.run to execute the command.
        result = subprocess.run(
            data.command,
            shell=True,
            cwd=self.project_root,
            capture_output=True,
            text=True,
        )
        logger.info(f"Command executed with return code: {result.returncode}")
        if result.stdout:
            logger.debug(f"stdout: {result.stdout}")
        if result.stderr:
            logger.error(f"stderr: {result.stderr}")
        return {
            "stdout": result.stdout,
            "stderr": result.stderr,
            "returncode": result.returncode,
        }

    def create_directory(self, data: CreateDirectory):
        """
        Create a new directory relative to the project root.
        """
        dir_path = self.project_root / data.directory_path
        logger.debug(f"Creating directory at: {dir_path}")
        dir_path.mkdir(parents=True, exist_ok=True)
        logger.info(f"Directory created: {dir_path}")
        return f"Directory created: {dir_path}"

    def delete_file(self, data: DeleteFile):
        """
        Delete a file relative to the project root.
        """
        file_path = self.project_root / data.file_path
        logger.debug(f"Attempting to delete file at: {file_path}")
        if file_path.exists():
            file_path.unlink()
            logger.info(f"File deleted: {file_path}")
            return f"File deleted: {file_path}"
        else:
            logger.warning(f"File not found: {file_path}")
            return f"File not found: {file_path}"
