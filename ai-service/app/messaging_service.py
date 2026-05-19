import json
import redis

from app.config import settings


class MessagingService:
    def __init__(self):
        try:
            self.redis_client = redis.Redis(
                host=settings.REDIS_HOST,
                port=settings.REDIS_PORT,
                decode_responses=True
            )

            self.redis_client.ping()
            self.redis_available = True

        except Exception:
            self.redis_client = None
            self.redis_available = False

    def publish_ai_message(
        self,
        game_id: str,
        message: str,
        message_type: str = "AI_MESSAGE",
        audio_url: str | None = None
    ) -> None:
        payload = {
            "game_id": game_id,
            "message_type": message_type,
            "message": message,
            "audio_url": audio_url
        }

        if not self.redis_available:
            print("Redis is not available. Message was not published.")
            print(payload)
            return

        self.redis_client.publish(
            settings.GAME_EVENT_CHANNEL,
            json.dumps(payload)
        )