from pydantic import BaseModel


class VoiceConfig(BaseModel):
    id: str
    name: str
    locale: str  # e.g. en-US
    style: str   # neutral, dramatic, festive, etc.
