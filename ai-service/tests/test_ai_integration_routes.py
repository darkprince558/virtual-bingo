import unittest

import app.main as main_app
from app.ai_generation_service import (
    AIGenerationService,
    CallerAssetResponse,
    CallerAssetsBulkResponse,
)

app = main_app.app


class FailingCallerAssetProvider:
    def generate_caller_asset(self, request):
        raise RuntimeError("mock generation failed")


class MixedCallerAssetsBulkProvider:
    def generate_caller_assets_bulk(self, request):
        return CallerAssetsBulkResponse(
            gameRunId=request.game_run_id,
            provider="mixed-provider",
            assets=[
                CallerAssetResponse(
                    callDeckItemId=request.deck[0].call_deck_item_id,
                    word=request.deck[0].word,
                    sequence=request.deck[0].sequence,
                    line="Roadmap is up. Mark it if you have it!",
                    audioUrl="https://mock-storage.local/roadmap.mp3",
                    status="ready",
                    fallbackText="Next word is Roadmap.",
                    provider="mixed-provider"
                ),
                CallerAssetResponse(
                    callDeckItemId=request.deck[1].call_deck_item_id,
                    word=request.deck[1].word,
                    sequence=request.deck[1].sequence,
                    line="Next word is Launch.",
                    status="failed",
                    fallbackText="Next word is Launch.",
                    provider="mixed-provider",
                    providerMetadata={"type": "storage-fallback"}
                ),
                CallerAssetResponse(
                    callDeckItemId=request.deck[2].call_deck_item_id,
                    word=request.deck[2].word,
                    sequence=request.deck[2].sequence,
                    line="Focus means valid bingo.",
                    status="ready",
                    fallbackText="Next word is Focus.",
                    provider="mixed-provider"
                )
            ]
        )


class AIIntegrationRoutesTest(unittest.TestCase):
    def setUp(self):
        self.client = app.test_client()

    def test_health_check(self):
        response = self.client.get("/health")

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["status"], "healthy")
        self.assertIn("service", payload)
        self.assertIn("version", payload)

    def test_game_prep_normal_request(self):
        response = self.client.post("/ai/v1/game-prep", json={
            "gameRunId": "game-1",
            "topicPrompt": "Launch Week",
            "wordCount": 24,
            "tone": "fun",
            "audience": "engineering",
            "excludedWords": [],
            "settings": {
                "callerStyle": "bright host",
                "themeMode": "workplace"
            }
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["topic"], "Launch Week")
        self.assertEqual(payload["summary"], "Deterministic mock bingo content for Launch Week.")
        self.assertEqual(payload["callerStyle"], "bright host")
        self.assertEqual(payload["themePrompt"], "Launch Week theme using workplace mode")
        self.assertGreaterEqual(len(payload["words"]), 24)
        self.assertNotIn("", payload["words"])
        self.assertEqual(len(payload["words"]), len(set(payload["words"])))
        self.assertEqual(payload["words"][0], "Launch Kickoff")

    def test_game_prep_excludes_requested_words(self):
        response = self.client.post("/ai/v1/game-prep", json={
            "gameRunId": "game-1",
            "topicPrompt": "Launch Week",
            "wordCount": 24,
            "tone": "fun",
            "audience": "engineering",
            "excludedWords": ["Launch Kickoff", "week planning"],
            "settings": {
                "callerStyle": "bright host",
                "themeMode": "workplace"
            }
        })

        self.assertEqual(response.status_code, 200)
        words = response.get_json()["words"]
        lowered = {word.lower() for word in words}
        self.assertNotIn("launch kickoff", lowered)
        self.assertNotIn("week planning", lowered)
        self.assertGreaterEqual(len(words), 24)
        self.assertNotIn("", words)
        self.assertEqual(len(words), len(set(words)))

    def test_game_prep_custom_word_count(self):
        response = self.client.post("/ai/v1/game-prep", json={
            "gameRunId": "game-1",
            "topicPrompt": "Quarterly Planning",
            "wordCount": 30,
            "tone": "professional",
            "audience": "leaders",
            "excludedWords": [],
            "settings": {
                "callerStyle": "focused host",
                "themeMode": "classic"
            }
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["topic"], "Quarterly Planning")
        self.assertEqual(payload["callerStyle"], "focused host")
        self.assertEqual(payload["themePrompt"], "Quarterly Planning theme using classic mode")
        self.assertGreaterEqual(len(payload["words"]), 30)
        self.assertEqual(len(payload["words"]), len(set(payload["words"])))

    def test_game_prep_missing_optional_settings(self):
        response = self.client.post("/ai/v1/game-prep", json={
            "gameRunId": "game-1",
            "topicPrompt": "Team Celebration",
            "wordCount": 24,
            "tone": "calm",
            "audience": "all hands",
            "excludedWords": []
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["topic"], "Team Celebration")
        self.assertEqual(payload["callerStyle"], "calm")
        self.assertEqual(payload["themePrompt"], "Team Celebration theme using mock mode")
        self.assertGreaterEqual(len(payload["words"]), 24)

    def test_caller_assets_bulk_multiple_deck_items(self):
        response = self.client.post("/ai/v1/caller-assets/bulk", json={
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": [
                {
                    "callDeckItemId": "deck-1",
                    "word": "Roadmap",
                    "sequence": 1
                },
                {
                    "callDeckItemId": "deck-2",
                    "word": "Launch",
                    "sequence": 7
                }
            ]
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["gameRunId"], "game-1")
        self.assertEqual(payload["provider"], "mock-python-ai-service")
        self.assertEqual(payload["providerMetadata"]["voiceName"], "mock-host")
        self.assertEqual(len(payload["assets"]), 2)
        self.assertEqual([asset["callDeckItemId"] for asset in payload["assets"]], ["deck-1", "deck-2"])
        self.assertEqual([asset["sequence"] for asset in payload["assets"]], [1, 7])
        self.assertEqual(payload["assets"][0]["line"], "Roadmap is up. Mark it if you have it!")
        self.assertEqual(payload["assets"][1]["line"], "Launch is up. Mark it if you have it!")
        self.assertEqual(payload["assets"][0]["audioUrl"], "https://mock-ai.local/audio/0001-deck-1.mp3")
        self.assertEqual(payload["assets"][1]["audioUrl"], "https://mock-ai.local/audio/0007-deck-2.mp3")
        self.assertEqual(payload["assets"][0]["fallbackText"], "Next word is Roadmap.")
        self.assertEqual(payload["assets"][1]["fallbackText"], "Next word is Launch.")
        for asset in payload["assets"]:
            self.assertEqual(asset["status"], "ready")
            self.assertEqual(asset["provider"], "mock-python-ai-service")
            self.assertEqual(asset["providerMetadata"]["voiceName"], "mock-host")

    def test_caller_assets_bulk_empty_deck(self):
        response = self.client.post("/ai/v1/caller-assets/bulk", json={
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": []
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["gameRunId"], "game-1")
        self.assertEqual(payload["assets"], [])
        self.assertEqual(payload["provider"], "mock-python-ai-service")

    def test_caller_assets_bulk_missing_voice_name_uses_fallback(self):
        response = self.client.post("/ai/v1/caller-assets/bulk", json={
            "gameRunId": "game-1",
            "tone": "calm",
            "deck": [
                {
                    "callDeckItemId": "deck-1",
                    "word": "Roadmap",
                    "sequence": 3
                }
            ]
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["providerMetadata"]["voiceName"], "mock-voice")
        self.assertEqual(payload["assets"][0]["providerMetadata"]["voiceName"], "mock-voice")
        self.assertEqual(payload["assets"][0]["line"], "Roadmap is next. Take a quick look.")
        self.assertEqual(payload["assets"][0]["sequence"], 3)

    def test_caller_assets_bulk_output_is_stable(self):
        request_payload = {
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": [
                {
                    "callDeckItemId": "deck-1",
                    "word": "Roadmap",
                    "sequence": 1
                }
            ]
        }

        first = self.client.post("/ai/v1/caller-assets/bulk", json=request_payload)
        second = self.client.post("/ai/v1/caller-assets/bulk", json=request_payload)

        self.assertEqual(first.status_code, 200)
        self.assertEqual(second.status_code, 200)
        self.assertEqual(first.get_json(), second.get_json())

    def test_caller_assets_bulk_returns_http_200_for_partial_failure(self):
        original_service = main_app.ai_generation_service
        main_app.ai_generation_service = AIGenerationService(
            caller_assets_bulk_provider=MixedCallerAssetsBulkProvider()
        )
        try:
            response = self.client.post("/ai/v1/caller-assets/bulk", json={
                "gameRunId": "game-1",
                "voiceName": "mock-host",
                "tone": "fun",
                "deck": [
                    {"callDeckItemId": "deck-1", "word": "Roadmap", "sequence": 1},
                    {"callDeckItemId": "deck-2", "word": "Launch", "sequence": 2},
                    {"callDeckItemId": "deck-3", "word": "Focus", "sequence": 3}
                ]
            })
        finally:
            main_app.ai_generation_service = original_service

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["gameRunId"], "game-1")
        self.assertEqual([asset["status"] for asset in payload["assets"]], ["ready", "failed", "fallback"])
        self.assertEqual(payload["assets"][0]["audioUrl"], "https://mock-storage.local/roadmap.mp3")
        self.assertEqual(payload["assets"][1]["fallbackText"], "Next word is Launch.")
        self.assertEqual(payload["assets"][2]["line"], "Next word is Focus.")

    def test_caller_asset_successful_single_asset(self):
        response = self.client.post("/ai/v1/caller-assets", json={
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 1,
            "voiceName": "local-default",
            "tone": "fun"
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["callDeckItemId"], "deck-1")
        self.assertEqual(payload["word"], "Roadmap")
        self.assertEqual(payload["sequence"], 1)
        self.assertEqual(payload["line"], "Roadmap is up. Mark it if you have it!")
        self.assertEqual(payload["audioUrl"], "https://mock-ai.local/audio/0001-deck-1.mp3")
        self.assertEqual(payload["status"], "ready")
        self.assertEqual(payload["fallbackText"], "Next word is Roadmap.")
        self.assertEqual(payload["provider"], "mock-python-ai-service")
        self.assertEqual(payload["providerMetadata"]["voiceName"], "local-default")

    def test_caller_asset_missing_optional_tone_uses_fallback_tone(self):
        response = self.client.post("/ai/v1/caller-assets", json={
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 4,
            "voiceName": "local-default"
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["line"], "Roadmap is next. Check your card.")
        self.assertEqual(payload["sequence"], 4)
        self.assertEqual(payload["providerMetadata"]["tone"], "neutral")

    def test_caller_asset_returns_safe_fallback_when_generation_fails(self):
        original_service = main_app.ai_generation_service
        main_app.ai_generation_service = AIGenerationService(
            caller_asset_provider=FailingCallerAssetProvider()
        )
        try:
            response = self.client.post("/ai/v1/caller-assets", json={
                "gameRunId": "game-1",
                "callDeckItemId": "deck-1",
                "word": "Roadmap",
                "sequence": 5,
                "voiceName": "local-default",
                "tone": "fun"
            })
        finally:
            main_app.ai_generation_service = original_service

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["callDeckItemId"], "deck-1")
        self.assertEqual(payload["word"], "Roadmap")
        self.assertEqual(payload["sequence"], 5)
        self.assertEqual(payload["line"], "Next word is Roadmap.")
        self.assertEqual(payload["status"], "failed")
        self.assertEqual(payload["fallbackText"], "Next word is Roadmap.")
        self.assertEqual(payload["provider"], "local-fallback")
        self.assertEqual(payload["providerMetadata"]["type"], "fallback")
        self.assertIn("mock generation failed", payload["providerMetadata"]["error"])

    def test_theme_generate_christmas_theme(self):
        response = self.client.post("/ai/v1/themes/generate", json={
            "gameRunId": "game-1",
            "prompt": "Christmas party",
            "tone": "cheerful",
            "allowedAssets": ["snowflake", "gift", "star", "briefcase"]
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["name"], "Christmas party")
        self.assertIn("festive", payload["summary"])
        self.assertEqual(payload["palette"]["accent"], "#B91C1C")
        self.assertEqual(payload["icons"], ["snowflake", "gift", "star"])
        self.assertEqual(payload["decorations"], ["snowflake", "gift", "star"])
        self.assertEqual(payload["motion"], "gentle")
        self.assertEqual(payload["callerTone"], "cheerful")
        self.assertTrue(payload["accessibility"]["contrastChecked"])
        self.assertTrue(payload["accessibility"]["avoidColorOnlyMeaning"])

    def test_theme_generate_workplace_theme(self):
        response = self.client.post("/ai/v1/themes/generate", json={
            "gameRunId": "game-1",
            "prompt": "workplace celebration",
            "tone": "professional",
            "allowedAssets": ["briefcase", "calendar", "clipboard", "sparkles"]
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["name"], "workplace celebration")
        self.assertIn("team-friendly", payload["summary"])
        self.assertEqual(payload["palette"]["primary"], "#1F7A8C")
        self.assertEqual(payload["icons"], ["briefcase", "calendar", "clipboard", "sparkles"])
        self.assertEqual(payload["decorations"], ["sparkles", "clipboard"])
        self.assertEqual(payload["motion"], "subtle")
        self.assertEqual(payload["callerTone"], "professional")

    def test_theme_generate_restricts_allowed_assets(self):
        response = self.client.post("/ai/v1/themes/generate", json={
            "gameRunId": "game-1",
            "prompt": "Christmas",
            "tone": "cheerful",
            "allowedAssets": ["gift"]
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["icons"], ["gift"])
        self.assertEqual(payload["decorations"], ["gift"])

    def test_theme_generate_has_no_unsafe_fields(self):
        response = self.client.post("/ai/v1/themes/generate", json={
            "gameRunId": "game-1",
            "prompt": "javascript:alert(1) url(https://bad.example/image.png)",
            "tone": "calm",
            "allowedAssets": [
                "sparkles",
                "https://bad.example/icon.png",
                "javascript:alert(1)",
                "<script>"
            ]
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        response_text = str(payload).lower()
        self.assertNotIn("javascript", response_text)
        self.assertNotIn("https://", response_text)
        self.assertNotIn("url(", response_text)
        self.assertNotIn("<script", response_text)
        self.assertEqual(payload["icons"], ["sparkles"])
        self.assertEqual(payload["decorations"], ["sparkles"])


if __name__ == "__main__":
    unittest.main()
