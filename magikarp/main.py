from fastapi import FastAPI

from magikarp.routers.chat import chat_router
from magikarp.routers.recommendations import recommendation_router
from magikarp.routers.notifications import notification_router

app = FastAPI()

# Router for dynamic pre-defined prompt recommendations
app.include_router(recommendation_router)
# Router for companion chats / general LLM usage
app.include_router(chat_router)
# Router for simulating notifications over the course of the day (2024-05-15)
app.include_router(notification_router)


@app.get("/")
async def read_root():
    return {
        "Hello": "it's me, Magikarp! Please add '/docs' to the URL in your browser to view and manually trigger my APIs."}
