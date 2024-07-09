import logging
from fastapi import Depends, APIRouter, HTTPException, status
from magikarp.services.model import TransformerModel
from magikarp.dependencies import get_transformer_service
from magikarp.models.requests import PromptRequest
from magikarp.models.responses import RecommendationResponse

# Configure the logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger('uvicorn.debug')

recommendation_router = APIRouter(prefix="/recommendation", tags=["recommendation"])

@recommendation_router.post("", response_model=RecommendationResponse, status_code=status.HTTP_200_OK)
async def ask_predefined_suggested_prompt(
        prompt_data: PromptRequest = Depends(),
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    """API endpoint used to ask the device predetermined suggested prompts. Such as, "what should I do today?" or,
    "do I have any meetings soon?"


    Args:
        prompt_data: The input data containing predefined prompts.
        transformer_service: The transformer service dependency to interact with the model.

    Returns:
        A dictionary with the model's suggested prompts.

    Raises:
        HTTPException: If the prompts are empty or an error occurs.
    """
    logger.info(f"Received predefined prompt data: {prompt_data}")

    if not prompt_data.prompts:
        logger.error("Prompts cannot be empty")
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Prompts cannot be empty.")

    try:
        response = transformer_service.ask_model(prompt_data.prompts)
        logger.info(f"Llama3 response: {response}")
        return {"recommendation": response}
    except Exception as e:
        logger.error(f"Error generating suggestions: {e}")
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail="Error generating suggestions")

@recommendation_router.post("/suggest", response_model=RecommendationResponse, status_code=status.HTTP_200_OK)
async def generate_suggested_prompts(
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    """Endpoint to generate suggested prompts to ask the model.

    Args:
        transformer_service: The transformer service dependency to interact with the model.

    Returns:
        A dictionary with the model's generated suggestions.

    Raises:
        HTTPException: If an error occurs while generating suggestions.
    """
    logger.info("Generating suggested prompts")

    try:
        suggested_prompts = transformer_service.generate_suggested_prompts()
        logger.info(f"Generated suggested prompts: {suggested_prompts}")
        return {"recommendations": suggested_prompts}
    except Exception as e:
        logger.error(f"Error generating suggested prompts: {e}")
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail="Error generating suggested prompts")
