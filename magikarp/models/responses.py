from pydantic import BaseModel, Field


class RecommendationResponse(BaseModel):
    """
    Pydantic model for the API response.

    Attributes:
        response (str): The generated recommendation/response based on the user prompt.
    """
    recommendation: str

class ChatResponse(BaseModel):
    """
    Pydantic model for the API response.

    Attributes:
        response (str): The generated recommendation/response based on the user prompt.
    """
    response: str = Field(..., example="Hey! It's me, Magikarp! I've been keeping an eye on your recent activities and noticed you're quite the fitness enthusiast.")

