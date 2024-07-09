from magikarp.services.model import TransformerModel
from fastapi import Depends, APIRouter
from magikarp.dependencies import get_transformer_service
from magikarp.models.responses import CompanionResponse

notification_router = APIRouter(prefix="/notification", tags=["notification"])


@notification_router.get("/", response_model=CompanionResponse)
async def get_push_notifications_for_the_day(
        time: str,
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    pass
    # if prompt:
    #     response = transformer_service.ask_model(prompt)
    #     return {"response": response}
