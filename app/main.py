from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.routers import chat_router, plugins
from app.plugins.manager import PluginManager
from app.services.model import TransformerModel
from app.dependencies import set_app_state
import os
from app.plugins.code_dump.main import ProjectScanner

# Create Plugin Manager instance
plugin_manager = PluginManager()

# Create shared TransformerModel instance with plugin manager
transformer_model = TransformerModel(plugin_manager=plugin_manager)

# Store the transformer_model in app state for dependency injection
set_app_state("transformer_model", transformer_model)

# Create FastAPI instance
app = FastAPI(
    title="Magikarp API",
    description="APIs for Magikarp service with plugin support.",
    version="1.0.0"
)

# CORS (Cross-Origin Resource Sharing) Middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    allow_headers=["*"],
)

# Include routers
app.include_router(chat_router)
app.include_router(plugins.router)  # Include the plugins router

# Root endpoint
@app.get("/")
async def read_root():
    """
    Root endpoint of the API.

    Returns:
        dict: A welcome message with instructions to access the documentation.
    """
    return {
        "message": "Hello! Welcome to the Magikarp API with plugin support.",
        "instructions": "Add '/docs' to the URL in your browser to view and manually trigger the APIs."
    }


scanner = ProjectScanner("/Users/pprunty/GitHub/clunk")
project_info = scanner.scan()  # Returns string with structure and contents

# Write to file
scanner.scan_to_file("output.txt")