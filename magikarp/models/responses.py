from pydantic import BaseModel


class RecommendationResponse(BaseModel):
    """
    Pydantic model for the API response.

    Attributes:
        response (str): The generated recommendation/response based on the user prompt.
    """
    recommendation: str


class CompanionResponse(BaseModel):
    """
    Pydantic model for the API response.

    Attributes:
        response (str): The generated recommendation/response based on the user prompt.
    """
    response: str
