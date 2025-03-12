# settings.py
from pydantic_settings import BaseSettings
from pathlib import Path

class Settings(BaseSettings):
    model: str = "magikarp"
    PROJECT_OUTLINE: str = "You are building a basic 'hello world' program in Rust. The program should print 'Hello, World!' to the console. All rust dependencies should be installed if not already."

    class Config:
        # Build the path to the .env file located one directory up from this file.
        env_file = str(Path(__file__).resolve().parent.parent / ".env")
        env_file_encoding = "utf-8"

# Instantiate settings to be used throughout your FastAPI project
settings = Settings()
