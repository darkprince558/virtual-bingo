import unittest

from app.ai_generation_service import (
    AIGenerationService,
    CallerAssetRequest,
    CallerAssetResponse,
    CallerAssetsBulkRequest,
    CallerAssetsBulkResponse,
    GamePrepRequest,
    GamePrepResponse,
    MockCallerAssetProvider,
    MockCallerAssetsBulkProvider,
    MockGamePrepProvider,
    MockThemeProvider,
    ThemeRequest,
    ThemeResponse,
    safe_caller_line,
)
from app.audio_storage_provider import MockStorageProvider
from app.speech_provider import MockSpeechProvider, SpeechResult


class ByteSpeechProvider:
    def synthesize(self, request):
        return SpeechResult(
            audio_bytes=f"audio:{request.text}".encode("utf-8"),
            storage_key=f"caller-assets/{request.sequence:04d}-{request.asset_id}.mp3",
            provider="byte-speech",
            metadata={
                "type": "mock-bytes",
                "voiceName": request.voice_name
            }
        )


class FailingSpeechProvider:
    def synthesize(self, request):
        raise RuntimeError("speech failed")


class StorageKeySpeechProvider:
    def synthesize(self, request):
        return SpeechResult(
            storage_key=f"storage/{request.sequence}-{request.asset_id}.mp3",
            provider="storage-speech",
            metadata={
                "type": "mock-storage",
                "voiceName": request.voice_name
            }
        )


class FailingStorageProvider:
    def __init__(self, fail_on_storage_key: str | None = None):
        self.fail_on_storage_key = fail_on_storage_key

    def store(self, request):
        if self.fail_on_storage_key is None or request.storage_key == self.fail_on_storage_key:
            raise RuntimeError(f"storage failed for {request.storage_key}")
        return MockStorageProvider(return_url=True).store(request)


class FakeGamePrepProvider:
    def __init__(self):
        self.request = None

    def generate_game_prep(self, request):
        self.request = request
        return GamePrepResponse(
            topic="Fake Topic",
            summary="Fake summary",
            words=["Fake Word"],
            callerStyle="fake caller",
            themePrompt="fake theme"
        )


class FakeCallerAssetsBulkProvider:
    def __init__(self):
        self.request = None

    def generate_caller_assets_bulk(self, request):
        self.request = request
        return CallerAssetsBulkResponse(
            gameRunId="game-1",
            assets=[],
            provider="fake-provider"
        )


class FakeCallerAssetProvider:
    def __init__(self):
        self.request = None

    def generate_caller_asset(self, request):
        self.request = request
        return CallerAssetResponse(
            callDeckItemId="deck-1",
            word="Fake Word",
            sequence=1,
            line="Fake line",
            status="ready",
            provider="fake-provider"
        )


class FailingCallerAssetProvider:
    def generate_caller_asset(self, request):
        raise RuntimeError("provider failed")


class UnsafeCallerAssetProvider:
    def generate_caller_asset(self, request):
        return CallerAssetResponse(
            callDeckItemId=request.call_deck_item_id,
            word=request.word,
            sequence=request.sequence,
            line=f"{request.word} means winner confirmed!",
            status="ready",
            provider="unsafe-provider"
        )


class UnsafeCallerAssetsBulkProvider:
    def generate_caller_assets_bulk(self, request):
        return CallerAssetsBulkResponse(
            gameRunId=request.game_run_id,
            assets=[
                CallerAssetResponse(
                    callDeckItemId=item.call_deck_item_id,
                    word=item.word,
                    sequence=item.sequence,
                    line=f"{item.word}: player has won.",
                    status="ready",
                    provider="unsafe-provider"
                )
                for item in request.deck
            ],
            provider="unsafe-provider"
        )


class UnsafeThemeProvider:
    def generate_theme(self, request):
        return ThemeResponse(
            name="Unsafe",
            summary="Unsafe structured theme",
            palette={
                "primary": "#111111",
                "badCss": "url(https://bad.example/image.png)",
                "script": "javascript:alert(1)"
            },
            icons=["sparkles", "https://bad.example/icon.png", "javascript:alert(1)"],
            decorations=["confetti", "<script>"],
            motion="subtle",
            callerTone="calm",
            accessibility={}
        )


class AIGenerationServiceTest(unittest.TestCase):
    def test_service_delegates_to_provider(self):
        provider = FakeGamePrepProvider()
        service = AIGenerationService(game_prep_provider=provider)
        request = GamePrepRequest.model_validate({
            "gameRunId": "game-1",
            "topicPrompt": "Launch Week",
            "wordCount": 1,
            "excludedWords": []
        })

        response = service.generate_game_prep(request)

        self.assertIs(provider.request, request)
        self.assertEqual(response.topic, "Fake Topic")
        self.assertEqual(response.words, ["Fake Word"])

    def test_mock_provider_returns_stable_game_prep(self):
        provider = MockGamePrepProvider()
        request = GamePrepRequest.model_validate({
            "gameRunId": "game-1",
            "topicPrompt": "Launch Week",
            "wordCount": 24,
            "tone": "fun",
            "excludedWords": ["Launch Kickoff"],
            "settings": {
                "callerStyle": "bright host",
                "themeMode": "workplace"
            }
        })

        first = provider.generate_game_prep(request)
        second = provider.generate_game_prep(request)

        self.assertEqual(first, second)
        self.assertEqual(first.topic, "Launch Week")
        self.assertEqual(first.summary, "Deterministic mock bingo content for Launch Week.")
        self.assertEqual(first.caller_style, "bright host")
        self.assertEqual(first.theme_prompt, "Launch Week theme using workplace mode")
        self.assertGreaterEqual(len(first.words), 24)
        self.assertNotIn("Launch Kickoff", first.words)
        self.assertEqual(len(first.words), len(set(first.words)))
        self.assertNotIn("", first.words)

    def test_service_delegates_caller_assets_bulk_to_provider(self):
        provider = FakeCallerAssetsBulkProvider()
        service = AIGenerationService(caller_assets_bulk_provider=provider)
        request = CallerAssetsBulkRequest.model_validate({
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": []
        })

        response = service.generate_caller_assets_bulk(request)

        self.assertIs(provider.request, request)
        self.assertEqual(response.game_run_id, "game-1")
        self.assertEqual(response.assets, [])

    def test_mock_caller_assets_bulk_provider_preserves_deck_order_and_sequence(self):
        provider = MockCallerAssetsBulkProvider(speech_provider=MockSpeechProvider())
        request = CallerAssetsBulkRequest.model_validate({
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": [
                {"callDeckItemId": "deck-1", "word": "Roadmap", "sequence": 2},
                {"callDeckItemId": "deck-2", "word": "Launch", "sequence": 9}
            ]
        })

        response = provider.generate_caller_assets_bulk(request)

        self.assertEqual(response.game_run_id, "game-1")
        self.assertEqual([asset.call_deck_item_id for asset in response.assets], ["deck-1", "deck-2"])
        self.assertEqual([asset.sequence for asset in response.assets], [2, 9])
        self.assertEqual(response.assets[0].audio_url, "https://mock-ai.local/audio/0002-deck-1.mp3")
        self.assertEqual(response.assets[1].audio_url, "https://mock-ai.local/audio/0009-deck-2.mp3")
        self.assertEqual(response.assets[0].fallback_text, "Next word is Roadmap.")
        self.assertEqual(response.provider_metadata["voiceName"], "mock-host")
        self.assertEqual(response.assets[0].provider_metadata["speechProvider"], "mock-speech")

    def test_service_delegates_caller_asset_to_provider(self):
        provider = FakeCallerAssetProvider()
        service = AIGenerationService(caller_asset_provider=provider)
        request = CallerAssetRequest.model_validate({
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 1,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        response = service.generate_caller_asset(request)

        self.assertIs(provider.request, request)
        self.assertEqual(response.provider, "fake-provider")
        self.assertEqual(response.line, "Fake line")

    def test_mock_caller_asset_provider_generates_single_asset(self):
        provider = MockCallerAssetProvider(speech_provider=MockSpeechProvider())
        request = CallerAssetRequest.model_validate({
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 6,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        response = provider.generate_caller_asset(request)

        self.assertEqual(response.call_deck_item_id, "deck-1")
        self.assertEqual(response.word, "Roadmap")
        self.assertEqual(response.sequence, 6)
        self.assertEqual(response.line, "Roadmap is up. Mark it if you have it!")
        self.assertEqual(response.audio_url, "https://mock-ai.local/audio/0006-deck-1.mp3")
        self.assertEqual(response.status, "ready")
        self.assertEqual(response.fallback_text, "Next word is Roadmap.")
        self.assertEqual(response.provider_metadata["voiceName"], "mock-host")
        self.assertEqual(response.provider_metadata["speechProvider"], "mock-speech")

    def test_single_caller_asset_can_return_storage_key_from_speech_provider(self):
        provider = MockCallerAssetProvider(speech_provider=StorageKeySpeechProvider())
        request = CallerAssetRequest.model_validate({
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 6,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        response = provider.generate_caller_asset(request)

        self.assertIsNone(response.audio_url)
        self.assertEqual(response.storage_key, "storage/6-deck-1.mp3")
        self.assertEqual(response.status, "ready")
        self.assertEqual(response.provider_metadata["speechProvider"], "storage-speech")

    def test_single_caller_asset_stores_speech_bytes_successfully(self):
        provider = MockCallerAssetProvider(
            speech_provider=ByteSpeechProvider(),
            storage_provider=MockStorageProvider(return_url=False)
        )
        request = CallerAssetRequest.model_validate({
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 6,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        response = provider.generate_caller_asset(request)

        self.assertIsNone(response.audio_url)
        self.assertEqual(response.storage_key, "caller-assets/0006-deck-1.mp3")
        self.assertEqual(response.status, "ready")
        self.assertEqual(response.provider_metadata["speechProvider"], "byte-speech")
        self.assertEqual(response.provider_metadata["storage"]["provider"], "mock-storage")
        self.assertTrue(response.provider_metadata["storage"]["stored"])

    def test_single_caller_asset_falls_back_when_storage_fails(self):
        provider = MockCallerAssetProvider(
            speech_provider=ByteSpeechProvider(),
            storage_provider=FailingStorageProvider()
        )
        request = CallerAssetRequest.model_validate({
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 6,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        response = provider.generate_caller_asset(request)

        self.assertEqual(response.line, "Next word is Roadmap.")
        self.assertEqual(response.status, "failed")
        self.assertEqual(response.fallback_text, "Next word is Roadmap.")
        self.assertIsNone(response.audio_url)
        self.assertIsNone(response.storage_key)
        self.assertEqual(response.provider_metadata["type"], "storage-fallback")
        self.assertIn("storage failed", response.provider_metadata["error"])

    def test_bulk_caller_assets_allow_partial_storage_success(self):
        provider = MockCallerAssetsBulkProvider(
            speech_provider=ByteSpeechProvider(),
            storage_provider=FailingStorageProvider(
                fail_on_storage_key="caller-assets/0009-deck-2.mp3"
            )
        )
        request = CallerAssetsBulkRequest.model_validate({
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": [
                {"callDeckItemId": "deck-1", "word": "Roadmap", "sequence": 2},
                {"callDeckItemId": "deck-2", "word": "Launch", "sequence": 9}
            ]
        })

        response = provider.generate_caller_assets_bulk(request)

        self.assertEqual(len(response.assets), 2)
        self.assertEqual(response.assets[0].status, "ready")
        self.assertEqual(response.assets[0].audio_url, "https://mock-storage.local/caller-assets/0002-deck-1.mp3")
        self.assertIsNone(response.assets[0].storage_key)
        self.assertEqual(response.assets[1].status, "failed")
        self.assertEqual(response.assets[1].line, "Next word is Launch.")
        self.assertEqual(response.assets[1].fallback_text, "Next word is Launch.")
        self.assertIsNone(response.assets[1].audio_url)
        self.assertIsNone(response.assets[1].storage_key)
        self.assertEqual(response.assets[1].provider_metadata["type"], "storage-fallback")

    def test_single_caller_asset_falls_back_when_speech_provider_fails(self):
        provider = MockCallerAssetProvider(speech_provider=FailingSpeechProvider())
        request = CallerAssetRequest.model_validate({
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 6,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        response = provider.generate_caller_asset(request)

        self.assertEqual(response.line, "Next word is Roadmap.")
        self.assertEqual(response.status, "failed")
        self.assertEqual(response.fallback_text, "Next word is Roadmap.")
        self.assertIsNone(response.audio_url)
        self.assertIsNone(response.storage_key)
        self.assertIn("speech failed", response.provider_metadata["error"])

    def test_bulk_caller_assets_fall_back_per_item_when_speech_provider_fails(self):
        provider = MockCallerAssetsBulkProvider(speech_provider=FailingSpeechProvider())
        request = CallerAssetsBulkRequest.model_validate({
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": [
                {"callDeckItemId": "deck-1", "word": "Roadmap", "sequence": 2}
            ]
        })

        response = provider.generate_caller_assets_bulk(request)

        self.assertEqual(response.assets[0].line, "Next word is Roadmap.")
        self.assertEqual(response.assets[0].status, "failed")
        self.assertEqual(response.assets[0].fallback_text, "Next word is Roadmap.")
        self.assertIn("speech failed", response.assets[0].provider_metadata["error"])

    def test_service_returns_safe_fallback_when_single_asset_provider_fails(self):
        service = AIGenerationService(caller_asset_provider=FailingCallerAssetProvider())
        request = CallerAssetRequest.model_validate({
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 6,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        response = service.generate_caller_asset(request)

        self.assertEqual(response.call_deck_item_id, "deck-1")
        self.assertEqual(response.word, "Roadmap")
        self.assertEqual(response.sequence, 6)
        self.assertEqual(response.line, "Next word is Roadmap.")
        self.assertEqual(response.status, "failed")
        self.assertEqual(response.fallback_text, "Next word is Roadmap.")
        self.assertEqual(response.provider, "local-fallback")
        self.assertIn("provider failed", response.provider_metadata["error"])

    def test_blocked_single_caller_line_is_replaced_with_fallback(self):
        service = AIGenerationService(caller_asset_provider=UnsafeCallerAssetProvider())
        request = CallerAssetRequest.model_validate({
            "gameRunId": "game-1",
            "callDeckItemId": "deck-1",
            "word": "Roadmap",
            "sequence": 6,
            "voiceName": "mock-host",
            "tone": "fun"
        })

        response = service.generate_caller_asset(request)

        self.assertEqual(response.line, "Next word is Roadmap.")
        self.assertEqual(response.fallback_text, "Next word is Roadmap.")
        self.assertEqual(response.status, "fallback")
        self.assertEqual(response.provider, "unsafe-provider")

    def test_blocked_bulk_caller_line_is_replaced_with_fallback(self):
        service = AIGenerationService(
            caller_assets_bulk_provider=UnsafeCallerAssetsBulkProvider()
        )
        request = CallerAssetsBulkRequest.model_validate({
            "gameRunId": "game-1",
            "voiceName": "mock-host",
            "tone": "fun",
            "deck": [
                {"callDeckItemId": "deck-1", "word": "Roadmap", "sequence": 2}
            ]
        })

        response = service.generate_caller_assets_bulk(request)

        self.assertEqual(response.assets[0].line, "Next word is Roadmap.")
        self.assertEqual(response.assets[0].fallback_text, "Next word is Roadmap.")
        self.assertEqual(response.assets[0].status, "fallback")
        self.assertEqual(response.assets[0].provider, "unsafe-provider")

    def test_all_blocked_phrases_are_filtered(self):
        blocked_phrases = [
            "winner confirmed",
            "player has won",
            "valid bingo",
            "invalid bingo",
            "game result",
            "the game is over",
            "official winner",
            "confirmed winner",
        ]

        for phrase in blocked_phrases:
            with self.subTest(phrase=phrase):
                line = safe_caller_line(
                    line=f"Roadmap means {phrase}.",
                    word="Roadmap"
                )
                self.assertEqual(line, "Next word is Roadmap.")

    def test_mock_theme_provider_generates_christmas_tokens(self):
        provider = MockThemeProvider()
        request = ThemeRequest.model_validate({
            "gameRunId": "game-1",
            "prompt": "Christmas party",
            "tone": "cheerful",
            "allowedAssets": ["snowflake", "gift", "star", "briefcase"]
        })

        response = provider.generate_theme(request)

        self.assertEqual(response.name, "Christmas party")
        self.assertEqual(response.icons, ["snowflake", "gift", "star"])
        self.assertEqual(response.decorations, ["snowflake", "gift", "star"])
        self.assertTrue(response.accessibility["contrastChecked"])
        self.assertTrue(response.accessibility["avoidColorOnlyMeaning"])

    def test_theme_service_sanitizes_unsafe_provider_output(self):
        service = AIGenerationService(theme_provider=UnsafeThemeProvider())
        request = ThemeRequest.model_validate({
            "gameRunId": "game-1",
            "prompt": "Unsafe",
            "tone": "calm",
            "allowedAssets": ["sparkles", "confetti"]
        })

        response = service.generate_theme(request)

        self.assertEqual(response.icons, ["sparkles"])
        self.assertEqual(response.decorations, ["confetti"])
        self.assertEqual(response.palette, {"primary": "#111111"})
        self.assertTrue(response.accessibility["contrastChecked"])
        self.assertTrue(response.accessibility["avoidColorOnlyMeaning"])


if __name__ == "__main__":
    unittest.main()
