import html
from dataclasses import dataclass, field
from typing import Protocol

import httpx

from app.config import settings


@dataclass
class SpeechRequest:
    text: str
    voice_name: str
    asset_id: str
    sequence: int


@dataclass
class SpeechResult:
    audio_url: str | None = None
    storage_key: str | None = None
    audio_bytes: bytes | None = None
    content_type: str = "audio/mpeg"
    provider: str = "unknown"
    metadata: dict[str, object] = field(default_factory=dict)


class SpeechProvider(Protocol):
    def synthesize(self, request: SpeechRequest) -> SpeechResult:
        ...


class MockSpeechProvider:
    provider_name = "mock-speech"

    def synthesize(self, request: SpeechRequest) -> SpeechResult:
        return SpeechResult(
            audio_url=mock_audio_url(request.asset_id, request.sequence),
            provider=self.provider_name,
            metadata={
                "type": "mock",
                "voiceName": request.voice_name,
                "contentType": "audio/mpeg"
            }
        )


class AzureSpeechProvider:
    provider_name = "azure-speech"

    def __init__(
        self,
        endpoint: str | None = None,
        key: str | None = None,
        region: str | None = None,
        default_voice_name: str | None = None,
        timeout_seconds: float | None = None
    ):
        self.key = key if key is not None else settings.AZURE_SPEECH_KEY
        self.region = region if region is not None else settings.AZURE_SPEECH_REGION
        self.endpoint = endpoint if endpoint is not None else settings.AZURE_SPEECH_ENDPOINT
        self.default_voice_name = (
            default_voice_name
            if default_voice_name is not None
            else settings.DEFAULT_VOICE_NAME
        )
        self.timeout_seconds = timeout_seconds or settings.AZURE_SPEECH_TIMEOUT_SECONDS

    def synthesize(self, request: SpeechRequest) -> SpeechResult:
        if not self.key:
            raise RuntimeError("Azure Speech key is missing")

        endpoint = self._endpoint()
        voice_name = request.voice_name or self.default_voice_name
        response = httpx.post(
            endpoint,
            headers={
                "Ocp-Apim-Subscription-Key": self.key,
                "Content-Type": "application/ssml+xml",
                "X-Microsoft-OutputFormat": "audio-24khz-48kbitrate-mono-mp3",
                "User-Agent": "bingo-ai-narration-service"
            },
            content=self._ssml(text=request.text, voice_name=voice_name),
            timeout=self.timeout_seconds
        )
        response.raise_for_status()
        audio_bytes = response.content
        return SpeechResult(
            audio_bytes=audio_bytes,
            storage_key=storage_key(request.asset_id, request.sequence),
            provider=self.provider_name,
            metadata={
                "type": "azure",
                "voiceName": voice_name,
                "region": self.region,
                "contentType": "audio/mpeg",
                "byteLength": len(audio_bytes)
            }
        )

    def _endpoint(self) -> str:
        if self.endpoint:
            return self.endpoint.rstrip("/")
        if not self.region:
            raise RuntimeError("Azure Speech region is missing")
        return f"https://{self.region}.tts.speech.microsoft.com/cognitiveservices/v1"

    def _ssml(self, text: str, voice_name: str) -> bytes:
        escaped_text = html.escape(text)
        escaped_voice = html.escape(voice_name, quote=True)
        return (
            '<speak version="1.0" xml:lang="en-US">'
            f'<voice xml:lang="en-US" name="{escaped_voice}">'
            f"{escaped_text}"
            "</voice>"
            "</speak>"
        ).encode("utf-8")


def default_speech_provider() -> SpeechProvider:
    if settings.TTS_PROVIDER.lower() == "azure":
        return AzureSpeechProvider()
    return MockSpeechProvider()


def mock_audio_url(asset_id: str, sequence: int) -> str:
    return f"https://mock-ai.local/audio/{sequence:04d}-{safe_asset_id(asset_id)}.mp3"


def storage_key(asset_id: str, sequence: int) -> str:
    return f"caller-assets/{sequence:04d}-{safe_asset_id(asset_id)}.mp3"


def safe_asset_id(asset_id: str) -> str:
    safe_id = "".join(
        character if character.isalnum() or character in "-_" else "-"
        for character in asset_id.strip()
    )
    return safe_id or "asset"
