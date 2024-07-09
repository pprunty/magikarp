from fastapi import Depends, APIRouter, HTTPException, status
from magikarp.services.model import TransformerModel
from magikarp.dependencies import get_transformer_service
from magikarp.models.responses import ChatResponse
from magikarp.models.requests import DateRequest, AddRuleRequest

notification_router = APIRouter(prefix="/notification", tags=["notification"])

@notification_router.post("/get", response_model=ChatResponse, status_code=status.HTTP_200_OK)
async def get_push_notifications_for_the_day(
        request: DateRequest,
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    """API endpoint to get push notifications for the day based on user-defined rules.

    Args:
        request (DateRequest): The request body containing the date and rules.
        transformer_service (TransformerModel): The transformer service dependency.

    Returns:
        dict: The model's response containing push notifications.
    """
    try:
        response = transformer_service.get_push_notifications(request.request_date, request.rules)
        return {"response": response}
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(e))


@notification_router.post("/add_rule", status_code=status.HTTP_201_CREATED)
async def add_new_rule(
        request: AddRuleRequest,
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    """API endpoint to add a new rule to the rule set.

    Args:
        request (AddRuleRequest): The request body containing the new rule description.
        transformer_service (TransformerModel): The transformer service dependency.

    Returns:
        dict: A confirmation message.
    """
    try:
        transformer_service.add_rule(request.rule_description)
        return {"message": "New rule added successfully"}
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(e))
