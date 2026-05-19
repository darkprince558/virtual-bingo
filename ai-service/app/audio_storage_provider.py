from dataclasses import dataclass, field
from typing import Protocol

from azure.storage.blob import BlobServiceClient, ContentSettings

from app.config import settings


@dataclass
class AudioStorageRequest:
    audio_bytes: bytes
    storage_key: str
    content_type: str = "audio/mpeg"


@dataclass
class AudioStorageResult:
    audio_url: str | None = None
    storage_key: str | None = None
    provider: str = "unknown"
    metadata: dict[str, object] = field(default_factory=dict)


class AudioStorageProvider(Protocol):
    def store(self, request: AudioStorageRequest) -> AudioStorageResult:
        ...


class MockStorageProvider:
    provider_name = "mock-storage"

    def __init__(self, return_url: bool = True):
        self.return_url = return_url

    def store(self, request: AudioStorageRequest) -> AudioStorageResult:
        if self.return_url:
            return AudioStorageResult(
                audio_url=f"https://mock-storage.local/{request.storage_key}",
                provider=self.provider_name,
                metadata={
                    "type": "mock",
                    "contentType": request.content_type,
                    "byteLength": len(request.audio_bytes)
                }
            )

        return AudioStorageResult(
            storage_key=request.storage_key,
            provider=self.provider_name,
            metadata={
                "type": "mock",
                "contentType": request.content_type,
                "byteLength": len(request.audio_bytes)
            }
        )


class AzureBlobStorageProvider:
    provider_name = "azure-blob-storage"

    def __init__(
        self,
        connection_string: str | None = None,
        container_name: str | None = None
    ):
        self.connection_string = (
            connection_string
            if connection_string is not None
            else settings.AZURE_STORAGE_CONNECTION_STRING
        )
        self.container_name = (
            container_name
            if container_name is not None
            else settings.AZURE_STORAGE_CONTAINER
        )
        self.blob_service_client = (
            BlobServiceClient.from_connection_string(self.connection_string)
            if self.connection_string
            else None
        )

    def store(self, request: AudioStorageRequest) -> AudioStorageResult:
        if not self.blob_service_client:
            raise RuntimeError("Azure Storage connection string is missing")

        blob_client = self.blob_service_client.get_blob_client(
            container=self.container_name,
            blob=request.storage_key
        )
        blob_client.upload_blob(
            request.audio_bytes,
            overwrite=True,
            content_settings=ContentSettings(content_type=request.content_type)
        )
        return AudioStorageResult(
            audio_url=blob_client.url,
            storage_key=request.storage_key,
            provider=self.provider_name,
            metadata={
                "type": "azure",
                "container": self.container_name,
                "contentType": request.content_type,
                "byteLength": len(request.audio_bytes)
            }
        )


def default_audio_storage_provider() -> AudioStorageProvider:
    if settings.AUDIO_STORAGE_PROVIDER.lower() == "azure":
        return AzureBlobStorageProvider()
    return MockStorageProvider()
