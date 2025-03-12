from typing import List, Union
from pydantic import BaseModel, Field

class ChatRequest(BaseModel):
    """Simple pydantic model to store prompt to LLM"""
    prompt: str = Field(..., example="Hello there.")

class UpdateFile(BaseModel):
    """Model for creating a file"""
    file_name: str = Field(..., example="main.cpp")
    file_content: str = Field(
        ...,
        example="#include <iostream>\n\nint main() {\n    std::cout << \"Hello, World!\";\n    return 0;\n}"
    )
    file_path: str = Field(..., example="~/Desktop/")
    new_file: bool = Field(..., example=True)

class CommandRequest(BaseModel):
    """Model for running a CLI command"""
    command: str = Field(..., example="ls -l")

class CreateDirectory(BaseModel):
    """Model for creating a directory"""
    directory_path: str = Field(..., example="~/rust-basic/src")

class DeleteFile(BaseModel):
    """Model for deleting a file"""
    file_path: str = Field(..., example="~/rust-basic/main.rs")

class InstallDependencies(BaseModel):
    """Model for installing dependencies"""
    package_manager: str = Field(..., example="cargo")
    dependencies: List[str] = Field(..., example=["serde", "tokio"])

class RemoveDependencies(BaseModel):
    """Model for removing dependencies"""
    package_manager: str = Field(..., example="cargo")
    dependencies: List[str] = Field(..., example=["serde"])

class RunProject(BaseModel):
    """Model for running the project"""
    command: str = Field(..., example="cargo run")

class BuildProject(BaseModel):
    """Model for building the project"""
    command: str = Field(..., example="cargo build")

class RunTests(BaseModel):
    """Model for running tests"""
    command: str = Field(..., example="cargo test")

class InitializeGit(BaseModel):
    """Model for initializing a git repository"""
    command: str = Field(..., example="git init")

class CommitChanges(BaseModel):
    """Model for committing changes to git"""
    commit_message: str = Field(..., example="Initial commit")
    commands: List[str] = Field(..., example=["git add .", "git commit -m 'Initial commit'"])

class CreateConfigFile(BaseModel):
    """Model for creating a configuration file (e.g., Cargo.toml, .gitignore)"""
    file_name: str = Field(..., example="Cargo.toml")
    file_content: str = Field(
        ...,
        example="[package]\nname = \"rust-basic\"\nversion = \"0.1.0\"\nedition = \"2021\""
    )
    file_path: str = Field(..., example="~/rust-basic/")

class StartServer(BaseModel):
    """Model for starting a development server"""
    command: str = Field(..., example="cargo run")

class DeployProject(BaseModel):
    """Model for deploying the project"""
    command: str = Field(..., example="scp -r rust-basic user@server:/var/www/")

# JSON Schema Model that encapsulates all possible actions and their corresponding payloads

class ActionResponse(BaseModel):
    """
    JSON schema for the action response.

    This model specifies the action to perform along with the necessary data payload.
    Valid actions include:
    - create_file
    - run_command
    - create_directory
    - delete_file
    - install_dependencies
    - remove_dependencies
    - run_project
    - build_project
    - run_tests
    - initialize_git
    - commit_changes
    - create_config_file
    - start_server
    - deploy_project
    """
    action: str = Field(
        ...,
        example="create_file",
        description="Specifies the type of action to be performed. "
                    "Options: create_file, run_command, create_directory, delete_file, "
                    "install_dependencies, remove_dependencies, run_project, build_project, "
                    "run_tests, initialize_git, commit_changes, create_config_file, start_server, deploy_project"
    )
    data: Union[
        UpdateFile,
        CommandRequest,
        CreateDirectory,
        DeleteFile,
        InstallDependencies,
        RemoveDependencies,
        RunProject,
        BuildProject,
        RunTests,
        InitializeGit,
        CommitChanges,
        CreateConfigFile,
        StartServer,
        DeployProject,
    ] = Field(..., description="The payload corresponding to the specified action.")
