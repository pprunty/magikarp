from magikarp.services.model import TransformerModel
from fastapi import Depends, APIRouter
from magikarp.dependencies import get_transformer_service
from magikarp.models.requests import PromptRequest
from magikarp.models.responses import RecommendationResponse

recommendation_router = APIRouter(prefix="/recommendation", tags=["recommendation"])


@recommendation_router.get("/", response_model=RecommendationResponse)
async def ask_predefined_suggested_prompt(
        prompt_data: PromptRequest = Depends(),
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    suggested_prompts = transformer_service.ask_model(prompt_data.prompts)

    return {"recommendations": suggested_prompts}


@recommendation_router.get("/suggest", response_model=RecommendationResponse)
async def generate_suggested_prompts(
        prompt_data: PromptRequest = Depends(),
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    suggested_prompts = transformer_service.generate_suggested_prompts()

    return {"recommendations": suggested_prompts}
