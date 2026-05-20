import unittest

import app.main as main_app
from app.ai_generation_service import (
    AIGenerationService,
    CallerAssetResponse,
    CallerAssetsBulkResponse,
)


app = main_app.app


FORBIDDEN_GAME_TRUTH_KEYS = {
    "officialNextWord",
    "nextWord",
    "cardState",
    "cardMarks",
    "validBingo",
    "invalidBingo",
    "claimValid",
    "claimInvalid",
    "winner",
    "winners",
    "officialWinner",
    "playerAccess",
    "gameStatus",
    "reward",
    "rewards",
    "auditOutcome",
}

FORBIDDEN_GAME_TRUTH_PHRASES = [
    "winner confirmed",
    "player has won",
    "valid bingo",
    "invalid bingo",
    "game result",
    "the game is over",
    "official winner",
    "confirmed winner",
    "official next word",
    "card state",
    "player access",
    "game status",
    "reward approved",
    "audit outcome",
]


class FailingCallerAssetProvider:
    def generate_caller_asset(self, request):
        raise RuntimeError("contract fallback failure")


class PartialCallerAssetsBulkProvider:
    def generate_caller_assets_bulk(self, request):
        return CallerAssetsBulkResponse(
            gameRunId=request.game_run_id,
            provider="contract-provider",
            assets=[
                CallerAssetResponse(
                    callDeckItemId=request.deck[0].call_deck_item_id,
                    word=request.deck[0].word,
                    sequence=request.deck[0].sequence,
                    line="Roadmap is up. Mark it if you have it!",
                    audioUrl="https://mock-storage.local/roadmap.mp3",
                    status="ready",
                    fallbackText="Next word is Roadmap.",
                    provider="contract-provider"
                ),
                CallerAssetResponse(
                    callDeckItemId=request.deck[1].call_deck_item_id,
                    word=request.deck[1].word,
                    sequence=request.deck[1].sequence,
                    line="Next word is Launch.",
                    status="failed",
                    fallbackText="Next word is Launch.",
                    provider="contract-provider"
                )
            ]
        )


class AIContractTest(unittest.TestCase):
    def setUp(self):
        self.client = app.test_client()

    def tearDown(self):
        main_app.ai_generation_service = AIGenerationService()

    def test_game_prep_contract(self):
        request = {
            "gameRunId": "game-1",
            "topicPrompt": "Launch Week",
            "wordCount": 24,
            "tone": "fun",
            "audience": "engineering",
            "excludedWords": ["Launch Kickoff"],
            "settings": {
                "callerStyle": "bright host",
                "themeMode": "workplace"
            }
        }

        first = self.client.post("/ai/v1/game-prep", json=request)
        second = self.client.post("/ai/v1/game-prep", json=request)

        self.assertEqual(first.status_code, 200)
        self.assertEqual(first.get_json(), second.get_json())
        payload = first.get_json()
        self.assert_required_fields(payload, [
            "topic",
            "summary",
            "words",
            "callerStyle",
            "themePrompt",
        ])
        self.assertGreaterEqual(len(payload["words"]), 24)
        self.assertNotIn("Launch Kickoff", payload["words"])
        self.assert_no_game_truth(payload)

    def test_game_prep_validates_request(self):
        response = self.client.post("/ai/v1/game-prep", json={
            "topicPrompt": "Launch Week",
            "wordCount": 24
        })

        self.assertEqual(response.status_code, 400)
        self.assertIn("error", response.get_json())

    def test_caller_assets_bulk_contract(self):
        request = {
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": [
                {"callDeckItemId": "deck-1", "word": "Roadmap", "sequence": 1},
                {"callDeckItemId": "deck-2", "word": "Launch", "sequence": 7}
            ]
        }

        first = self.client.post("/ai/v1/caller-assets/bulk", json=request)
        second = self.client.post("/ai/v1/caller-assets/bulk", json=request)

        self.assertEqual(first.status_code, 200)
        self.assertEqual(first.get_json(), second.get_json())
        payload = first.get_json()
        self.assert_required_fields(payload, ["gameRunId", "assets", "provider"])
        self.assertEqual(payload["gameRunId"], "game-1")
        self.assertEqual(len(payload["assets"]), 2)
        self.assertEqual([asset["sequence"] for asset in payload["assets"]], [1, 7])
        for asset in payload["assets"]:
            self.assert_caller_asset_shape(asset)
            self.assertIn(asset["status"], ["ready", "failed", "fallback"])
        self.assert_no_game_truth(payload)

    def test_caller_assets_bulk_validates_request(self):
        response = self.client.post("/ai/v1/caller-assets/bulk", json={
            "gameRunId": "game-1",
            "deck": [
                {"word": "Roadmap", "sequence": 1}
            ]
        })

        self.assertEqual(response.status_code, 400)
        self.assertIn("error", response.get_json())

    def test_caller_assets_bulk_partial_failure_contract(self):
        main_app.ai_generation_service = AIGenerationService(
            caller_assets_bulk_provider=PartialCallerAssetsBulkProvider()
        )
        response = self.client.post("/ai/v1/caller-assets/bulk", json={
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": [
                {"callDeckItemId": "deck-1", "word": "Roadmap", "sequence": 1},
                {"callDeckItemId": "deck-2", "word": "Launch", "sequence": 2}
            ]
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual([asset["status"] for asset in payload["assets"]], ["ready", "failed"])
        self.assertEqual(payload["assets"][1]["fallbackText"], "Next word is Launch.")
        self.assert_no_game_truth(payload)

    def test_caller_asset_contract(self):
        request = {
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 1,
            "voiceName": "mock-host",
            "tone": "fun"
        }

        first = self.client.post("/ai/v1/caller-assets", json=request)
        second = self.client.post("/ai/v1/caller-assets", json=request)

        self.assertEqual(first.status_code, 200)
        self.assertEqual(first.get_json(), second.get_json())
        payload = first.get_json()
        self.assert_caller_asset_shape(payload)
        self.assertEqual(payload["sequence"], 1)
        self.assertIn("Roadmap", payload["line"])
        self.assert_no_game_truth(payload)

    def test_caller_asset_validates_request(self):
        response = self.client.post("/ai/v1/caller-assets", json={
            "gameRunId": "game-1",
            "word": "Roadmap",
            "sequence": 1
        })

        self.assertEqual(response.status_code, 400)
        self.assertIn("error", response.get_json())

    def test_caller_asset_safe_fallback_contract(self):
        main_app.ai_generation_service = AIGenerationService(
            caller_asset_provider=FailingCallerAssetProvider()
        )
        response = self.client.post("/ai/v1/caller-assets", json={
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 1,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        self.assertEqual(response.status_code, 200)
        payload = response.get_json()
        self.assertEqual(payload["status"], "failed")
        self.assertEqual(payload["line"], "Next word is Roadmap.")
        self.assertEqual(payload["fallbackText"], "Next word is Roadmap.")
        self.assert_no_game_truth(payload)

    def test_theme_generate_contract(self):
        request = {
            "gameRunId": "game-1",
            "prompt": "workplace celebration",
            "tone": "professional",
            "allowedAssets": ["briefcase", "calendar", "clipboard", "sparkles"]
        }

        first = self.client.post("/ai/v1/themes/generate", json=request)
        second = self.client.post("/ai/v1/themes/generate", json=request)

        self.assertEqual(first.status_code, 200)
        self.assertEqual(first.get_json(), second.get_json())
        payload = first.get_json()
        self.assert_required_fields(payload, [
            "name",
            "summary",
            "palette",
            "icons",
            "decorations",
            "motion",
            "callerTone",
            "accessibility",
        ])
        self.assertEqual(set(payload["icons"]) - set(request["allowedAssets"]), set())
        self.assertEqual(set(payload["decorations"]) - set(request["allowedAssets"]), set())
        self.assertTrue(payload["accessibility"]["contrastChecked"])
        self.assertTrue(payload["accessibility"]["avoidColorOnlyMeaning"])
        self.assert_no_unsafe_theme_content(payload)
        self.assert_no_game_truth(payload)

    def test_theme_generate_validates_request(self):
        response = self.client.post("/ai/v1/themes/generate", json={
            "gameRunId": "game-1",
            "allowedAssets": ["sparkles"]
        })

        self.assertEqual(response.status_code, 400)
        self.assertIn("error", response.get_json())

    def assert_required_fields(self, payload, fields):
        for field in fields:
            self.assertIn(field, payload)
            self.assertIsNotNone(payload[field])

    def assert_caller_asset_shape(self, payload):
        self.assert_required_fields(payload, [
            "callDeckItemId",
            "word",
            "sequence",
            "line",
            "status",
            "provider",
        ])
        self.assertIn("fallbackText", payload)
        self.assertTrue(payload.get("audioUrl") or payload.get("storageKey") or payload["status"] in ["failed", "fallback"])

    def assert_no_game_truth(self, payload):
        keys = set(self.collect_keys(payload))
        self.assertEqual(keys & FORBIDDEN_GAME_TRUTH_KEYS, set())
        lowered = str(payload).lower()
        for phrase in FORBIDDEN_GAME_TRUTH_PHRASES:
            self.assertNotIn(phrase, lowered)

    def assert_no_unsafe_theme_content(self, payload):
        lowered = str(payload).lower()
        self.assertNotIn("javascript", lowered)
        self.assertNotIn("<script", lowered)
        self.assertNotIn("url(", lowered)
        self.assertNotIn("http://", lowered)
        self.assertNotIn("https://", lowered)

    def collect_keys(self, value):
        if isinstance(value, dict):
            for key, child in value.items():
                yield key
                yield from self.collect_keys(child)
        elif isinstance(value, list):
            for child in value:
                yield from self.collect_keys(child)


if __name__ == "__main__":
    unittest.main()
