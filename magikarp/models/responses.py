from typing import Dict, Union

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


class NotificationResponse(BaseModel):
    """
    Pydantic model for the API response.

    Attributes:
        notifications (Dict[str, str]): A dictionary where the keys are the notification times and the values are the notification messages.
    """
    notifications: Union[Dict[str, str], str] = Field(
        ...,
        example={
            "2024-05-15 08:30": "Hey Dan, don't forget our meeting at 9 AM today! Let's catch up beforehand.",
            "2024-05-15 09:00": "Time to get moving! You've got a team meeting in an hour. Don't forget to log your steps.",
            "2024-05-15 10:15": "Hey Dan, can you please review the latest code changes by 4 PM? Let's make sure we're on track for today's project planning.",
            "2024-05-15 11:00": "You've got a meeting in an hour! Take a few minutes to stretch and get settled before it starts. See you there!",
            "2024-05-15 14:00": "Client call in an hour! Make sure you're prepared with all the necessary materials. You got this!",
            "2024-05-15 16:00": "Code review time! Take a few minutes to grab a snack and get focused before diving back into your work.",
            "2024-05-15 17:00": "Hey Dan, don't forget to collect your running shoes from the store near work. You've earned it after today's workout!",
            "2024-05-15 18:30": "Time for a relaxing evening! Why not try listening to some new music? I recommend checking out Nicolas Jaar's latest album.",
            "2024-05-15 20:00": "Hey Dan, don't forget to log your daily steps and sleep goals. You're doing great so far this week!",
            "2024-05-15 21:30": "It's time for some social interaction! Why not reach out to a friend or family member to catch up? You've got nothing to lose!"
        }
    )