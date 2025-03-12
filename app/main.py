from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.routers import chat_router

# Create FastAPI instance
app = FastAPI(
    title="Magikarp API",
    description="APIs for Magikarp service.",
    version="1.0.0"
)

# CORS (Cross-Origin Resource Sharing) Middleware
# Adjust origins as per your deployment needs
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    allow_headers=["*"],
)

# Include routers
app.include_router(chat_router)


# Root endpoint
@app.get("/")
async def read_root():
    """
    Root endpoint of the API.

    Returns:
        dict: A welcome message with instructions to access the documentation.
    """
    return {
        "message": "Hello! Welcome to the Magikarp API.",
        "instructions": "Add '/docs' to the URL in your browser to view and manually trigger the APIs."
    }
