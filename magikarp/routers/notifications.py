import json
from fastapi import Depends, APIRouter, HTTPException, status
from magikarp.services.model import TransformerModel
from magikarp.dependencies import get_transformer_service
from magikarp.models.responses import NotificationResponse
from magikarp.models.requests import DateRequest

notification_router = APIRouter(prefix="/notification", tags=["notification"])


@notification_router.post("/get", response_model=NotificationResponse, status_code=status.HTTP_200_OK)
async def get_push_notifications_for_the_day(
        request: DateRequest,
        transformer_service: TransformerModel = Depends(get_transformer_service)
):
    """API endpoint to get push notifications for the day based on user-defined rules.

    Args:
        request (DateRequest): The request body containing the date and rules.
        transformer_service (TransformerModel): The transformer service dependency.

    Returns:
        NotificationResponse: The model's response containing push notifications.
    """
    try:
        response = transformer_service.get_push_notifications(request.request_date, request.rules)
        print(f"response = {response}")

        # Assuming the response contains JSON string after the initial message
        try:
            # Find the start of the JSON part
            start_index = response.index("{")
            json_str = response[start_index:]

            # Ensure the JSON string ends with a closing brace
            if not json_str.endswith("}"):
                json_str += "}"

            print(f"json_str = {json_str}")
            response_dict = json.loads(json_str)
            print(f"response dict = {response_dict}")
            return NotificationResponse(notifications=response_dict)
        except (ValueError, json.JSONDecodeError) as e:
            print(f"JSON parsing error: {e}")
            # If there's an error parsing the JSON, return the raw response as a string
            return NotificationResponse(notifications=response)
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(e))
