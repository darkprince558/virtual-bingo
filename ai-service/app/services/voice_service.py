from app.constants.voices import VOICES, DEFAULT_VOICE


class VoiceService:

    def resolve_voice(self, requested_voice_id: str | None):

        if not requested_voice_id:
            return DEFAULT_VOICE

        for v in VOICES:
            if v["id"] == requested_voice_id:
                return v

        return DEFAULT_VOICE
