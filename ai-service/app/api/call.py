from fastapi import APIRouter
from app.services.voice_service import VoiceService

router = APIRouter()
voice_service = VoiceService()


@router.post("/call")
async def call_number(payload: dict):

    number = payload["number"]
    theme = payload["theme"]
    voice_id = payload.get("voice_id")  # optional

    voice = voice_service.resolve_voice(voice_id)

    prompt = build_call_prompt(
        theme=theme,
        number=number,
        voice_style=voice["style"]
    )

    # LLM call happens here (existing logic)
    text = await generate_call(prompt)

    return {
        "number": number,
        "text": text,
        "voice": voice
    }
