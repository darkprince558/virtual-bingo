import uuid
from azure.storage.blob import BlobServiceClient, ContentSettings
from app.config import settings


class BlobStorageService:
    def __init__(self):
        self.connection_string = settings.AZURE_STORAGE_CONNECTION_STRING
        self.container_name = settings.AZURE_STORAGE_CONTAINER

        if self.connection_string:
            self.blob_service_client = BlobServiceClient.from_connection_string(
                self.connection_string
            )
        else:
            self.blob_service_client = None

    def upload_audio(self, audio_bytes: bytes) -> str:
        if not self.blob_service_client:
            raise RuntimeError(
                "Azure Storage connection string is missing. "
                "Set AZURE_STORAGE_CONNECTION_STRING in your .env file."
            )

        file_name = f"{uuid.uuid4()}.mp3"

        blob_client = self.blob_service_client.get_blob_client(
            container=self.container_name,
            blob=file_name
        )

        blob_client.upload_blob(
            audio_bytes,
            overwrite=True,
            content_settings=ContentSettings(content_type="audio/mpeg")
        )

        return blob_client.url