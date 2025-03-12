import json
import logging
import platform  # Import platform to get OS details
from pathlib import Path

# Configure logging for the SnapshotManager module.
logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
handler = logging.StreamHandler()
formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
handler.setFormatter(formatter)
logger.addHandler(handler)

class SnapshotManager:
    def __init__(self, project_root: str):
        """
        Args:
            project_root (str): The root directory of the project.
        """
        self.project_root = Path(project_root).expanduser().resolve()
        logger.info(f"Initialized SnapshotManager with project root: {self.project_root}")

    def get_snapshot(self) -> str:
        """
        Recursively traverses the project directory and collects all non-hidden files and their contents.

        Returns:
            A JSON string representing a snapshot including the OS type, and a list of files with their paths and contents.
            Example:
            {
              "os": "Windows",
              "files": [
                { "file": "dir1/file1.txt", "content": "..." },
                { "file": "dir2/file2.txt", "content": "..." }
              ]
            }
        """
        logger.info("Generating project snapshot as JSON.")
        files_list = self._traverse_tree_to_list(self.project_root)
        os_info = platform.system()  # Get the OS type (Windows, Linux, Darwin for macOS)
        snapshot = {
            "os": os_info,
            "files": files_list
        }
        logger.debug(f"Snapshot generated with OS: {os_info}")
        return json.dumps(snapshot, indent=2)

    def _traverse_tree_to_list(self, path: Path, base: Path = None) -> list:
        """
        Recursively traverses the directory tree and collects non-hidden files.

        Args:
            path (Path): The current directory or file path.
            base (Path): The base path to compute relative paths. Defaults to the project root.

        Returns:
            A list of dictionaries, each with 'file' and 'content' keys.
        """
        if base is None:
            base = self.project_root
        files_snapshot = []
        if path.is_dir():
            # Skip hidden directories
            if path.name.startswith('.') and path != self.project_root:
                logger.debug(f"Skipping hidden directory: {path}")
                return files_snapshot
            for child in sorted(path.iterdir(), key=lambda x: x.name):
                # Skip hidden files and directories
                if child.name.startswith('.'):
                    logger.debug(f"Skipping hidden file or directory: {child}")
                    continue
                files_snapshot.extend(self._traverse_tree_to_list(child, base))
        else:
            try:
                with open(path, 'r', encoding="utf-8") as f:
                    content = f.read().strip()
            except Exception as e:
                content = f"Error reading file: {e}"
                logger.error(f"Failed to read file {path}: {e}")
            relative_path = str(path.relative_to(base))
            files_snapshot.append({
                "file": relative_path,
                "content": content
            })
            logger.debug(f"Added file to snapshot: {relative_path}")
        return files_snapshot
