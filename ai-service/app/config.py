import os

try:
    from dotenv import load_dotenv
except ImportError:
    def load_dotenv() -> None:
        return None


load_dotenv()


VALID_AI_PROVIDER_MODES = {"mock", "real"}
VALID_TTS_PROVIDERS = {"mock", "azure"}
VALID_STORAGE_PROVIDERS = {"mock", "azure"}
VALID_LOG_LEVELS = {"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}


class Settings:
    def __init__(self):
        self.APP_NAME = get_env("APP_NAME", "Bingo AI Narration Service")
        self.APP_VERSION = get_env("APP_VERSION", "1.0.0")
        self.LOG_LEVEL = get_env("LOG_LEVEL", "INFO").upper()

        self.AI_PROVIDER_MODE = get_env("AI_PROVIDER_MODE", "mock").lower()

        self.REDIS_HOST = get_env("REDIS_HOST", "localhost")
        self.REDIS_PORT = parse_int_env("REDIS_PORT", 6379)
        self.GAME_EVENT_CHANNEL = get_env("GAME_EVENT_CHANNEL", "bingo:ai-messages")

        default_provider = "azure" if self.AI_PROVIDER_MODE == "real" else "mock"
        self.TTS_PROVIDER = get_env("TTS_PROVIDER", default_provider).lower()
        self.AUDIO_STORAGE_PROVIDER = get_env(
            "AUDIO_STORAGE_PROVIDER",
            default_provider
        ).lower()

        self.AZURE_SPEECH_KEY = get_env("AZURE_SPEECH_KEY", "")
        self.AZURE_SPEECH_REGION = get_env("AZURE_SPEECH_REGION", "canadacentral")
        self.AZURE_SPEECH_ENDPOINT = get_env("AZURE_SPEECH_ENDPOINT", "")
        self.AZURE_SPEECH_TIMEOUT_SECONDS = parse_float_env(
            "AZURE_SPEECH_TIMEOUT_SECONDS",
            30.0
        )

        self.AZURE_STORAGE_CONNECTION_STRING = get_env(
            "AZURE_STORAGE_CONNECTION_STRING",
            ""
        )
        self.AZURE_STORAGE_CONTAINER = get_env(
            "AZURE_STORAGE_CONTAINER",
            "bingo-narration-audio"
        )

        self.DEFAULT_VOICE_NAME = get_env(
            "DEFAULT_VOICE_NAME",
            "en-US-JennyNeural"
        )

        self.CORS_ALLOWED_ORIGINS = split_csv_env("CORS_ALLOWED_ORIGINS")
        self.SERVICE_TO_SERVICE_TOKEN = get_env("SERVICE_TO_SERVICE_TOKEN", "")

        self.validate()

    def validate(self) -> None:
        if self.AI_PROVIDER_MODE not in VALID_AI_PROVIDER_MODES:
            raise ValueError("AI_PROVIDER_MODE must be mock or real")
        if self.TTS_PROVIDER not in VALID_TTS_PROVIDERS:
            raise ValueError("TTS_PROVIDER must be mock or azure")
        if self.AUDIO_STORAGE_PROVIDER not in VALID_STORAGE_PROVIDERS:
            raise ValueError("AUDIO_STORAGE_PROVIDER must be mock or azure")
        if self.LOG_LEVEL not in VALID_LOG_LEVELS:
            raise ValueError("LOG_LEVEL must be DEBUG, INFO, WARNING, ERROR, or CRITICAL")
        if self.REDIS_PORT < 1 or self.REDIS_PORT > 65535:
            raise ValueError("REDIS_PORT must be between 1 and 65535")
        if self.AZURE_SPEECH_TIMEOUT_SECONDS <= 0:
            raise ValueError("AZURE_SPEECH_TIMEOUT_SECONDS must be positive")
        if self.TTS_PROVIDER == "azure":
            if not self.AZURE_SPEECH_KEY:
                raise ValueError("AZURE_SPEECH_KEY is required when TTS_PROVIDER=azure")
            if not self.AZURE_SPEECH_ENDPOINT and not self.AZURE_SPEECH_REGION:
                raise ValueError(
                    "AZURE_SPEECH_ENDPOINT or AZURE_SPEECH_REGION is required when TTS_PROVIDER=azure"
                )
        if self.AUDIO_STORAGE_PROVIDER == "azure":
            if not self.AZURE_STORAGE_CONNECTION_STRING:
                raise ValueError(
                    "AZURE_STORAGE_CONNECTION_STRING is required when AUDIO_STORAGE_PROVIDER=azure"
                )
            if not self.AZURE_STORAGE_CONTAINER:
                raise ValueError(
                    "AZURE_STORAGE_CONTAINER is required when AUDIO_STORAGE_PROVIDER=azure"
                )


def get_env(key: str, fallback: str) -> str:
    value = os.getenv(key, "").strip()
    return value if value else fallback


def parse_int_env(key: str, fallback: int) -> int:
    value = os.getenv(key, "").strip()
    if not value:
        return fallback
    try:
        return int(value)
    except ValueError as error:
        raise ValueError(f"{key} must be an integer") from error


def parse_float_env(key: str, fallback: float) -> float:
    value = os.getenv(key, "").strip()
    if not value:
        return fallback
    try:
        return float(value)
    except ValueError as error:
        raise ValueError(f"{key} must be a number") from error


def split_csv_env(key: str) -> list[str]:
    value = os.getenv(key, "")
    return [
        part.strip()
        for part in value.split(",")
        if part.strip()
    ]


settings = Settings()
