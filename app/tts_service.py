import uuid
import html
import httpx
import asyncio

from app.config import settings
from app.blob_storage_service import BlobStorageService


class TextToSpeechService:
    def __init__(self):
        self.speech_key = settings.AZURE_SPEECH_KEY
        self.speech_region = settings.AZURE_SPEECH_REGION
        self.default_voice_name = settings.DEFAULT_VOICE_NAME
        self.tts_provider = settings.TTS_PROVIDER

        self.blob_storage_service = BlobStorageService()

    def generate_audio(
        self,
        text: str,
        voice_name: str | None = None
    ) -> str:
        selected_voice = voice_name or self.default_voice_name

        if self.tts_provider == "mock":
            return self._mock_audio_url()

        if self.tts_provider == "azure":
            audio_bytes = asyncio.run(
                self._generate_azure_speech(
                    text=text,
                    voice_name=selected_voice
                )
            )

            return self.blob_storage_service.upload_audio(audio_bytes)

        return self._mock_audio_url()

    def _mock_audio_url(self) -> str:
        audio_id = str(uuid.uuid4())
        return f"https://mock-storage.local/bingo-audio/{audio_id}.mp3"

    async def _generate_azure_speech(
        self,
        text: str,
        voice_name: str
    ) -> bytes:
        if not self.speech_key:
            raise RuntimeError(
                "Azure Speech key is missing. "
                "Set AZURE_SPEECH_KEY in your .env file."
            )

        endpoint = (
            f"https://{self.speech_region}.tts.speech.microsoft.com/"
            "cognitiveservices/v1"
        )

        escaped_text = html.escape(text)

        ssml = f"""
        <speak version="1.0" xml:lang="en-US">
            <voice xml:lang="en-US" name="{voice_name}">
                {escaped_text}
            </voice>
        </speak>
        """

        headers = {
            "Ocp-Apim-Subscription-Key": self.speech_key,
            "Content-Type": "application/ssml+xml",
            "X-Microsoft-OutputFormat": "audio-24khz-48kbitrate-mono-mp3",
            "User-Agent": "bingo-ai-narration-service"
        }

        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(
                endpoint,
                headers=headers,
                content=ssml.encode("utf-8")
            )

        response.raise_for_status()
        return response.content