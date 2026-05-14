import os
from dotenv import load_dotenv

load_dotenv()


class Settings:
    APP_NAME = "Bingo AI Narration Service"
    APP_VERSION = "1.0.0"

    REDIS_HOST = os.getenv("REDIS_HOST", "localhost")
    REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))
    GAME_EVENT_CHANNEL = os.getenv("GAME_EVENT_CHANNEL", "bingo:ai-messages")

    TTS_PROVIDER = os.getenv("TTS_PROVIDER", "mock")

    AZURE_SPEECH_KEY = os.getenv("AZURE_SPEECH_KEY", "")
    AZURE_SPEECH_REGION = os.getenv("AZURE_SPEECH_REGION", "canadacentral")

    AZURE_STORAGE_CONNECTION_STRING = os.getenv(
        "AZURE_STORAGE_CONNECTION_STRING",
        ""
    )

    AZURE_STORAGE_CONTAINER = os.getenv(
        "AZURE_STORAGE_CONTAINER",
        "bingo-narration-audio"
    )

    DEFAULT_VOICE_NAME = os.getenv(
        "DEFAULT_VOICE_NAME",
        "en-US-JennyNeural"
    )


settings = Settings()