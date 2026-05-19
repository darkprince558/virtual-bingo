import unittest
from unittest.mock import patch

from app.config import Settings


class SettingsTest(unittest.TestCase):
    def test_mock_mode_defaults_are_safe_for_local_development(self):
        with patch.dict("os.environ", {}, clear=True):
            settings = Settings()

        self.assertEqual(settings.AI_PROVIDER_MODE, "mock")
        self.assertEqual(settings.TTS_PROVIDER, "mock")
        self.assertEqual(settings.AUDIO_STORAGE_PROVIDER, "mock")
        self.assertEqual(settings.LOG_LEVEL, "INFO")
        self.assertEqual(settings.DEFAULT_VOICE_NAME, "en-US-JennyNeural")

    def test_real_mode_uses_azure_defaults_when_credentials_are_configured(self):
        with patch.dict("os.environ", {
            "AI_PROVIDER_MODE": "real",
            "AZURE_SPEECH_KEY": "test-speech-key",
            "AZURE_SPEECH_REGION": "canadacentral",
            "AZURE_STORAGE_CONNECTION_STRING": "UseDevelopmentStorage=true",
            "AZURE_STORAGE_CONTAINER": "test-audio"
        }, clear=True):
            settings = Settings()

        self.assertEqual(settings.AI_PROVIDER_MODE, "real")
        self.assertEqual(settings.TTS_PROVIDER, "azure")
        self.assertEqual(settings.AUDIO_STORAGE_PROVIDER, "azure")

    def test_explicit_azure_speech_requires_key(self):
        with patch.dict("os.environ", {
            "TTS_PROVIDER": "azure",
            "AZURE_SPEECH_REGION": "canadacentral"
        }, clear=True):
            with self.assertRaisesRegex(ValueError, "AZURE_SPEECH_KEY"):
                Settings()

    def test_explicit_azure_storage_requires_connection_string(self):
        with patch.dict("os.environ", {
            "AUDIO_STORAGE_PROVIDER": "azure",
            "AZURE_STORAGE_CONTAINER": "test-audio"
        }, clear=True):
            with self.assertRaisesRegex(ValueError, "AZURE_STORAGE_CONNECTION_STRING"):
                Settings()

    def test_rejects_invalid_modes_and_log_levels(self):
        invalid_cases = [
            ({"AI_PROVIDER_MODE": "magic"}, "AI_PROVIDER_MODE"),
            ({"TTS_PROVIDER": "magic"}, "TTS_PROVIDER"),
            ({"AUDIO_STORAGE_PROVIDER": "magic"}, "AUDIO_STORAGE_PROVIDER"),
            ({"LOG_LEVEL": "LOUD"}, "LOG_LEVEL"),
        ]

        for env, expected_error in invalid_cases:
            with self.subTest(env=env):
                with patch.dict("os.environ", env, clear=True):
                    with self.assertRaisesRegex(ValueError, expected_error):
                        Settings()

    def test_parses_cors_allowed_origins(self):
        with patch.dict("os.environ", {
            "CORS_ALLOWED_ORIGINS": " http://localhost:3000,https://bingo.example ,, "
        }, clear=True):
            settings = Settings()

        self.assertEqual(
            settings.CORS_ALLOWED_ORIGINS,
            ["http://localhost:3000", "https://bingo.example"]
        )


if __name__ == "__main__":
    unittest.main()
