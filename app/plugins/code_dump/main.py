import os
import sys
from typing import List, Set, Optional
import fnmatch
from io import StringIO

class ProjectScanner:
    def __init__(self,
                 start_path: str,
                 file_types: Optional[List[str]] = None,
                 exclusion_file: Optional[str] = None,
                 use_gitignore: bool = True):  # Added parameter to control gitignore usage
        """
        Initialize the ProjectScanner with the given parameters.

        Args:
            start_path: The directory to scan
            file_types: Optional list of file extensions to include (e.g., ['.py', '.txt'])
            exclusion_file: Optional path to a file containing exclusion patterns
            use_gitignore: Whether to automatically use .gitignore patterns if found (default: True)
        """
        self.start_path = start_path
        self.file_types = file_types
        self.exclusion_patterns = set()

        # Load patterns from the exclusion file if provided
        if exclusion_file and os.path.exists(exclusion_file):
            self.exclusion_patterns.update(self._parse_exclusion_file(exclusion_file))

        # Check for and load .gitignore if enabled
        if use_gitignore:
            gitignore_path = os.path.join(start_path, '.gitignore')
            if os.path.exists(gitignore_path):
                self.exclusion_patterns.update(self._parse_exclusion_file(gitignore_path))

    def _parse_exclusion_file(self, file_path: str) -> Set[str]:
        """Parse the exclusion file (or .gitignore) and return a set of patterns."""
        patterns = set()
        if file_path and os.path.exists(file_path):
            with open(file_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    # Skip empty lines and comments
                    if not line or line.startswith('#'):
                        continue
                    # Handle negation (!) in gitignore - for simplicity, we're just ignoring these
                    if line.startswith('!'):
                        continue
                    # Add the pattern
                    patterns.add(line)
        return patterns

    def _is_excluded(self, path: str) -> bool:
        """Check if a path matches any exclusion pattern."""
        for pattern in self.exclusion_patterns:
            # Handle directory-specific patterns (ending with /)
            is_dir_pattern = pattern.endswith('/')
            if is_dir_pattern and not os.path.isdir(os.path.join(self.start_path, path)):
                continue

            # Remove trailing slash for matching
            if is_dir_pattern:
                pattern = pattern[:-1]

            # Handle patterns with directory separator
            if '/' in pattern:
                # Absolute pattern (starts with /)
                if pattern.startswith('/'):
                    pattern = pattern[1:]  # Remove leading slash
                    if path == pattern or path.startswith(pattern + os.sep):
                        return True
                # Relative pattern with directory separator
                else:
                    parts = pattern.split('/')
                    path_parts = path.split(os.sep)

                    # Try to match the pattern parts with the path parts
                    if len(parts) <= len(path_parts):
                        for i in range(len(path_parts) - len(parts) + 1):
                            match = True
                            for j in range(len(parts)):
                                if not fnmatch.fnmatch(path_parts[i + j], parts[j]):
                                    match = False
                                    break
                            if match:
                                return True
            # Simple wildcard pattern
            else:
                if fnmatch.fnmatch(os.path.basename(path), pattern):
                    return True

                # Check if any part of the path matches the pattern
                path_parts = path.split(os.sep)
                if any(fnmatch.fnmatch(part, pattern) for part in path_parts):
                    return True

        return False

    def _generate_directory_structure(self, dir_path: str, prefix: str = '') -> List[str]:
        """Generate the tree structure for a directory."""
        entries = os.listdir(dir_path)
        entries = sorted(entries, key=lambda x: (not os.path.isdir(os.path.join(dir_path, x)), x.lower()))
        tree = []
        for i, entry in enumerate(entries):
            rel_path = os.path.relpath(os.path.join(dir_path, entry), self.start_path)
            if self._is_excluded(rel_path):
                continue

            if i == len(entries) - 1:
                connector = '└── '
                new_prefix = prefix + '    '
            else:
                connector = '├── '
                new_prefix = prefix + '│   '

            full_path = os.path.join(dir_path, entry)
            if os.path.isdir(full_path):
                tree.append(f"{prefix}{connector}{entry}/")
                tree.extend(self._generate_directory_structure(full_path, new_prefix))
            else:
                tree.append(f"{prefix}{connector}{entry}")
        return tree

    def get_directory_structure(self) -> str:
        """Return the directory structure as a string."""
        tree = ['/ '] + self._generate_directory_structure(self.start_path)
        return '\n'.join(tree)

    def scan(self) -> str:
        """
        Scan the directory and return project structure and file contents as a string.

        Returns:
            A string containing the directory structure and file contents.
        """
        output = StringIO()

        # Write the directory structure
        output.write("Directory Structure:\n")
        output.write("-------------------\n")
        output.write(self.get_directory_structure())
        output.write("\n\n")
        output.write("File Contents:\n")
        output.write("--------------\n")

        for root, dirs, files in os.walk(self.start_path):
            rel_path = os.path.relpath(root, self.start_path)

            if self._is_excluded(rel_path):
                continue

            for file in files:
                file_rel_path = os.path.join(rel_path, file)
                if self._is_excluded(file_rel_path):
                    continue
                if self.file_types is None or any(file.endswith(ext) for ext in self.file_types):
                    file_path = os.path.join(root, file)

                    output.write(f"File: {file_rel_path}\n")
                    output.write("-" * 50 + "\n")

                    try:
                        with open(file_path, 'r', encoding='utf-8') as in_file:
                            content = in_file.read()
                            output.write(f"Content of {file_rel_path}:\n")
                            output.write(content)
                    except Exception as e:
                        output.write(f"Error reading file: {str(e)}. Content skipped.\n")

                    output.write("\n\n")

        return output.getvalue()

    def scan_to_file(self, output_file: str) -> None:
        """
        Scan the directory and write project structure and file contents to a file.

        Args:
            output_file: Path to the output file.
        """
        with open(output_file, 'w', encoding='utf-8') as out_file:
            out_file.write(self.scan())
        print(f"Scan complete. Results written to {output_file}")