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


app = Flask(__name__)

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


if __name__ == "__main__":
    app.run(debug=True, port=5000)
