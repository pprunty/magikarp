from magikarp.services.model import TransformerModel
from fastapi import Depends, APIRouter
from magikarp.dependencies import get_transformer_service
from magikarp.models.responses import CompanionResponse

chat_router = APIRouter(prefix="/chat", tags=["chat"])


@chat_router.get("", response_model=CompanionResponse)
async def chat(
        date: str,
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    """
    API endpoint to recommend user suggestions based on predefined prompts, or alternatively, a
    typed prompt by the user.

    Args:
        prompt_data (PromptsModel): The user prompted data.
        transformer_service (TransformerService): The transformer (LLM) service.

    Returns:
        RecommendationResponse: The generated recommendation.
        :param prompt:
    """

    if prompt:
        response = transformer_service.ask_model(prompt)
        return {"response": response}
