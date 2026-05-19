from flask import Flask, request, jsonify
from pydantic import BaseModel, ConfigDict, Field, ValidationError

from app.ai_generation_service import (
    AIGenerationService,
    CallerAssetRequest,
    CallerAssetsBulkRequest,
    GamePrepRequest,
    ThemeRequest,
)
from app.config import settings
from app.narration_service import NarrationService
from app.tts_service import TextToSpeechService
from app.messaging_service import MessagingService
from app.safety import validate_called_number


app = Flask(__name__)

narration_service = NarrationService()
tts_service = TextToSpeechService()
messaging_service = MessagingService()
ai_generation_service = AIGenerationService()


class APIModel(BaseModel):
    model_config = ConfigDict(populate_by_name=True)


def parse_request(model: type[BaseModel]):
    data = request.get_json(silent=True)
    if data is None:
        return None, (jsonify({"error": "Missing JSON request body."}), 400)

    try:
        return model.model_validate(data), None
    except ValidationError as error:
        return None, (jsonify({"error": "Invalid request body.", "details": error.errors()}), 400)


def jsonify_model(model: BaseModel):
    return jsonify(model.model_dump(by_alias=True, exclude_none=True))


@app.get("/")
def home():
    return jsonify({
        "message": "Bingo AI Narration Service is running"
    })


@app.get("/health")
def health_check():
    return jsonify({
        "status": "healthy",
        "service": settings.APP_NAME,
        "version": settings.APP_VERSION
    })


@app.post("/ai/v1/game-prep")
def ai_game_prep():
    payload, error_response = parse_request(GamePrepRequest)
    if error_response:
        return error_response

    return jsonify_model(ai_generation_service.generate_game_prep(payload))


@app.post("/ai/v1/caller-assets/bulk")
def ai_caller_assets_bulk():
    payload, error_response = parse_request(CallerAssetsBulkRequest)
    if error_response:
        return error_response

    return jsonify_model(ai_generation_service.generate_caller_assets_bulk(payload))


@app.post("/ai/v1/caller-assets")
def ai_caller_asset():
    payload, error_response = parse_request(CallerAssetRequest)
    if error_response:
        return error_response

    return jsonify_model(ai_generation_service.generate_caller_asset(payload))


@app.post("/ai/v1/themes/generate")
def ai_theme_generate():
    payload, error_response = parse_request(ThemeRequest)
    if error_response:
        return error_response

    return jsonify_model(ai_generation_service.generate_theme(payload))


@app.post("/narrate/call")
def narrate_bingo_call():
    data = request.get_json()

    if not data:
        return jsonify({
            "error": "Missing JSON request body."
        }), 400

    game_id = data.get("game_id")
    called_number = data.get("called_number")
    round_number = data.get("round_number")
    tone = data.get("tone", "fun")
    voice_name = data.get("voice_name", settings.DEFAULT_VOICE_NAME)

    if not game_id:
        return jsonify({"error": "game_id is required."}), 400

    if not called_number:
        return jsonify({"error": "called_number is required."}), 400

    if not validate_called_number(called_number):
        return jsonify({
            "error": "Invalid Bingo number format or range."
        }), 400

    try:
        narration_text = narration_service.generate_caller_script(
            called_number=called_number,
            tone=tone,
            round_number=round_number
        )

        audio_url = tts_service.generate_audio(
            text=narration_text,
            voice_name=voice_name
        )

        messaging_service.publish_ai_message(
            game_id=game_id,
            message=narration_text,
            message_type="AI_NARRATION",
            audio_url=audio_url
        )

        return jsonify({
            "game_id": game_id,
            "called_number": called_number,
            "narration_text": narration_text,
            "audio_url": audio_url,
            "message_type": "AI_NARRATION"
        })

    except Exception as error:
        return jsonify({
            "error": f"Failed to generate narration: {str(error)}"
        }), 500


@app.post("/describe")
def generate_description():
    data = request.get_json()

    if not data:
        return jsonify({
            "error": "Missing JSON request body."
        }), 400

    game_id = data.get("game_id")
    called_number = data.get("called_number")
    audience = data.get("audience", "general")

    if not game_id:
        return jsonify({"error": "game_id is required."}), 400

    if not called_number:
        return jsonify({"error": "called_number is required."}), 400

    if not validate_called_number(called_number):
        return jsonify({
            "error": "Invalid Bingo number format or range."
        }), 400

    description = narration_service.generate_description(
        called_number=called_number,
        audience=audience
    )

    messaging_service.publish_ai_message(
        game_id=game_id,
        message=description,
        message_type="AI_DESCRIPTION"
    )

    return jsonify({
        "game_id": game_id,
        "called_number": called_number,
        "description": description,
        "message_type": "AI_DESCRIPTION"
    })


@app.post("/host/assist")
def assist_host():
    data = request.get_json()

    if not data:
        return jsonify({
            "error": "Missing JSON request body."
        }), 400

    game_id = data.get("game_id")
    context = data.get("context", "")
    tone = data.get("tone", "professional")

    if not game_id:
        return jsonify({"error": "game_id is required."}), 400

    suggestion = narration_service.generate_host_commentary(
        context=context,
        tone=tone
    )

    return jsonify({
        "game_id": game_id,
        "suggestion": suggestion,
        "message_type": "HOST_ASSISTANT"
    })


@app.post("/messages/send")
def send_ai_message():
    data = request.get_json()

    if not data:
        return jsonify({
            "error": "Missing JSON request body."
        }), 400

    game_id = data.get("game_id")
    message = data.get("message")
    message_type = data.get("message_type", "AI_MESSAGE")
    audio_url = data.get("audio_url")

    if not game_id:
        return jsonify({"error": "game_id is required."}), 400

    if not message:
        return jsonify({"error": "message is required."}), 400

    messaging_service.publish_ai_message(
        game_id=game_id,
        message=message,
        message_type=message_type,
        audio_url=audio_url
    )

    return jsonify({
        "status": "sent",
        "game_id": game_id,
        "message_type": message_type
    })


if __name__ == "__main__":
    app.run(debug=True, port=5000)
