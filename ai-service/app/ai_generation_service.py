from typing import Protocol

from pydantic import BaseModel, ConfigDict, Field

from app.audio_storage_provider import (
    AudioStorageProvider,
    AudioStorageRequest,
    default_audio_storage_provider,
)
from app.safety import BANNED_AI_DECISION_PHRASES
from app.speech_provider import SpeechProvider, SpeechRequest, default_speech_provider


class APIModel(BaseModel):
    model_config = ConfigDict(populate_by_name=True)


class GamePrepSettings(APIModel):
    caller_style: str | None = Field(default=None, alias="callerStyle")
    theme_mode: str | None = Field(default=None, alias="themeMode")


class GamePrepRequest(APIModel):
    game_run_id: str = Field(alias="gameRunId")
    topic_prompt: str | None = Field(default=None, alias="topicPrompt")
    word_count: int = Field(default=24, alias="wordCount")
    tone: str | None = None
    audience: str | None = None
    excluded_words: list[str] = Field(default_factory=list, alias="excludedWords")
    settings: GamePrepSettings = Field(default_factory=GamePrepSettings)


class GamePrepResponse(APIModel):
    topic: str
    summary: str
    words: list[str]
    caller_style: str = Field(alias="callerStyle")
    theme_prompt: str = Field(alias="themePrompt")


class CallDeckItemRequest(APIModel):
    call_deck_item_id: str = Field(alias="callDeckItemId")
    word: str
    sequence: int


class CallerAssetsBulkRequest(APIModel):
    game_run_id: str = Field(alias="gameRunId")
    voice_name: str | None = Field(default=None, alias="voiceName")
    tone: str | None = None
    deck: list[CallDeckItemRequest]


class CallerAssetRequest(APIModel):
    game_run_id: str = Field(alias="gameRunId")
    call_deck_item_id: str = Field(alias="callDeckItemId")
    word: str
    sequence: int
    voice_name: str | None = Field(default=None, alias="voiceName")
    tone: str | None = None


class CallerAssetResponse(APIModel):
    call_deck_item_id: str = Field(alias="callDeckItemId")
    word: str
    sequence: int
    line: str
    audio_url: str | None = Field(default=None, alias="audioUrl")
    storage_key: str | None = Field(default=None, alias="storageKey")
    status: str
    fallback_text: str | None = Field(default=None, alias="fallbackText")
    provider: str
    provider_metadata: dict[str, object] = Field(default_factory=dict, alias="providerMetadata")


class CallerAssetsBulkResponse(APIModel):
    game_run_id: str = Field(alias="gameRunId")
    assets: list[CallerAssetResponse]
    provider: str
    provider_metadata: dict[str, object] = Field(default_factory=dict, alias="providerMetadata")


class ThemeRequest(APIModel):
    game_run_id: str | None = Field(default=None, alias="gameRunId")
    prompt: str
    tone: str | None = None
    allowed_assets: list[str] = Field(default_factory=list, alias="allowedAssets")


class ThemeResponse(APIModel):
    name: str
    summary: str
    palette: dict[str, str]
    icons: list[str]
    decorations: list[str]
    motion: str
    caller_tone: str = Field(alias="callerTone")
    accessibility: dict[str, bool]


class GamePrepProvider(Protocol):
    def generate_game_prep(self, request: GamePrepRequest) -> GamePrepResponse:
        ...


class CallerAssetsBulkProvider(Protocol):
    def generate_caller_assets_bulk(
        self,
        request: CallerAssetsBulkRequest
    ) -> CallerAssetsBulkResponse:
        ...


class CallerAssetProvider(Protocol):
    def generate_caller_asset(
        self,
        request: CallerAssetRequest
    ) -> CallerAssetResponse:
        ...


class ThemeProvider(Protocol):
    def generate_theme(self, request: ThemeRequest) -> ThemeResponse:
        ...


class MockGamePrepProvider:
    def generate_game_prep(self, request: GamePrepRequest) -> GamePrepResponse:
        topic = self._normalize_topic(request.topic_prompt)
        caller_style = (
            request.settings.caller_style
            or request.tone
            or "light workplace caller"
        )
        theme_mode = request.settings.theme_mode or "mock"

        return GamePrepResponse(
            topic=topic,
            summary=f"Deterministic mock bingo content for {topic}.",
            words=self._generate_words(
                topic=topic,
                word_count=request.word_count,
                excluded_words=request.excluded_words
            ),
            callerStyle=caller_style,
            themePrompt=f"{topic} theme using {theme_mode} mode"
        )

    def _normalize_topic(self, topic_prompt: str | None) -> str:
        topic = " ".join((topic_prompt or "").split())
        return topic or "General Workplace Bingo"

    def _topic_terms(self, topic: str) -> list[str]:
        terms = []
        for raw in topic.replace("-", " ").split():
            term = "".join(character for character in raw if character.isalnum())
            if term:
                terms.append(term.title())
        return terms or ["Team"]

    def _generate_words(
        self,
        topic: str,
        word_count: int,
        excluded_words: list[str]
    ) -> list[str]:
        requested_count = max(word_count, 1)
        excluded = {word.strip().lower() for word in excluded_words if word.strip()}
        terms = self._topic_terms(topic)
        suffixes = [
            "Kickoff",
            "Planning",
            "Milestone",
            "Standup",
            "Retrospective",
            "Roadmap",
            "Launch",
            "Focus",
            "Review",
            "Collaboration",
            "Workshop",
            "Demo",
            "Feedback",
            "Dashboard",
            "Strategy",
            "Update",
            "Insight",
            "Priority",
            "Win",
            "Checklist",
            "Handoff",
            "Delivery",
            "Support",
            "Celebrate",
            "Learning",
            "Alignment",
            "Sprint",
            "Goal",
            "Briefing",
            "Quality",
        ]

        words = []
        seen = set()
        index = 0
        max_attempts = requested_count + len(suffixes) * max(len(terms), 1) + 50
        while len(words) < requested_count and index < max_attempts:
            term = terms[index % len(terms)]
            suffix = suffixes[index % len(suffixes)]
            cycle = index // (len(terms) * len(suffixes))
            candidate = (
                f"{term} {suffix}"
                if cycle == 0
                else f"{term} {suffix} {cycle + 1}"
            )
            key = candidate.strip().lower()
            if candidate.strip() and key not in excluded and key not in seen:
                seen.add(key)
                words.append(candidate)
            index += 1

        return words


class MockThemeProvider:
    christmas_assets = ["snowflake", "gift", "star", "bell", "tree", "ornament"]
    workplace_assets = ["briefcase", "calendar", "clipboard", "coffee", "target", "sparkles"]
    default_assets = ["sparkles", "star", "confetti", "target"]

    def generate_theme(self, request: ThemeRequest) -> ThemeResponse:
        prompt = " ".join(request.prompt.split()) or "Bingo Theme"
        theme_kind = self._theme_kind(prompt)
        allowed_assets = self._allowed_assets(request.allowed_assets)
        token_set = self._token_set(theme_kind)
        icons = self._filter_assets(token_set["icons"], allowed_assets)
        decorations = self._filter_assets(token_set["decorations"], allowed_assets)
        return ThemeResponse(
            name=token_set["name"] if prompt.lower() in ["christmas", "workplace"] else prompt,
            summary=token_set["summary"],
            palette=token_set["palette"],
            icons=icons,
            decorations=decorations,
            motion=token_set["motion"],
            callerTone=self._caller_tone(request.tone, token_set["callerTone"]),
            accessibility={
                "contrastChecked": True,
                "avoidColorOnlyMeaning": True
            }
        )

    def _theme_kind(self, prompt: str) -> str:
        lowered = prompt.lower()
        if "christmas" in lowered or "holiday" in lowered:
            return "christmas"
        if "workplace" in lowered or "office" in lowered or "team" in lowered:
            return "workplace"
        return "default"

    def _allowed_assets(self, allowed_assets: list[str]) -> set[str]:
        return {
            asset.strip()
            for asset in allowed_assets
            if self._safe_asset_token(asset)
        }

    def _filter_assets(self, preferred_assets: list[str], allowed_assets: set[str]) -> list[str]:
        if not allowed_assets:
            return []
        return [
            asset
            for asset in preferred_assets
            if asset in allowed_assets
        ]

    def _safe_asset_token(self, asset: str) -> bool:
        clean_asset = asset.strip()
        lowered = clean_asset.lower()
        return (
            clean_asset
            and "/" not in clean_asset
            and "\\" not in clean_asset
            and ":" not in clean_asset
            and "<" not in clean_asset
            and ">" not in clean_asset
            and "url(" not in lowered
            and "javascript" not in lowered
            and lowered == clean_asset
        )

    def _token_set(self, theme_kind: str) -> dict[str, object]:
        if theme_kind == "christmas":
            return {
                "name": "Christmas Bingo",
                "summary": "A festive structured theme with warm seasonal tokens.",
                "palette": {
                    "primary": "#0F766E",
                    "accent": "#B91C1C",
                    "background": "#F8FAFC",
                    "surface": "#FFFFFF"
                },
                "icons": self.christmas_assets,
                "decorations": ["snowflake", "gift", "star"],
                "motion": "gentle",
                "callerTone": "cheerful holiday host"
            }
        if theme_kind == "workplace":
            return {
                "name": "Workplace Bingo",
                "summary": "A clean team-friendly theme using workplace tokens.",
                "palette": {
                    "primary": "#1F7A8C",
                    "accent": "#F25F5C",
                    "background": "#F7F9FB",
                    "surface": "#FFFFFF"
                },
                "icons": self.workplace_assets,
                "decorations": ["sparkles", "target", "clipboard"],
                "motion": "subtle",
                "callerTone": "upbeat workplace host"
            }
        return {
            "name": "Mock Bingo Theme",
            "summary": "A safe structured theme using only allowed tokens.",
            "palette": {
                "primary": "#1F7A8C",
                "accent": "#F25F5C",
                "background": "#F7F9FB",
                "surface": "#FFFFFF"
            },
            "icons": self.default_assets,
            "decorations": ["confetti", "sparkles", "star"],
            "motion": "subtle",
            "callerTone": "friendly host"
        }

    def _caller_tone(self, requested_tone: str | None, default_tone: str) -> str:
        clean_tone = " ".join((requested_tone or "").split())
        return clean_tone or default_tone


class MockCallerAssetsBulkProvider:
    provider_name = "mock-python-ai-service"
    default_voice_name = "mock-voice"

    def __init__(
        self,
        speech_provider: SpeechProvider | None = None,
        storage_provider: AudioStorageProvider | None = None
    ):
        self.speech_provider = speech_provider or default_speech_provider()
        self.storage_provider = storage_provider or default_audio_storage_provider()

    def generate_caller_assets_bulk(
        self,
        request: CallerAssetsBulkRequest
    ) -> CallerAssetsBulkResponse:
        voice_name = self._voice_name(request.voice_name)
        tone = self._tone(request.tone)
        return CallerAssetsBulkResponse(
            gameRunId=request.game_run_id,
            assets=[
                self._asset_for_item(
                    item=item,
                    voice_name=voice_name,
                    tone=tone
                )
                for item in request.deck
            ],
            provider=self.provider_name,
            providerMetadata={
                "type": "mock",
                "voiceName": voice_name,
                "tone": tone
            }
        )

    def _asset_for_item(
        self,
        item: CallDeckItemRequest,
        voice_name: str,
        tone: str
    ) -> CallerAssetResponse:
        clean_word = " ".join(item.word.split())
        fallback_text = fallback_caller_text(clean_word)
        line = build_caller_line(clean_word, tone)
        try:
            speech = self.speech_provider.synthesize(SpeechRequest(
                text=line,
                voice_name=voice_name,
                asset_id=item.call_deck_item_id,
                sequence=item.sequence
            ))
        except Exception as error:
            return failed_caller_asset(
                call_deck_item_id=item.call_deck_item_id,
                word=clean_word,
                sequence=item.sequence,
                fallback_text=fallback_text,
                provider=self.provider_name,
                error=error
            )

        try:
            audio_url, storage_key, storage_metadata = resolve_audio_output(
                speech=speech,
                storage_provider=self.storage_provider
            )
        except Exception as error:
            return failed_caller_asset(
                call_deck_item_id=item.call_deck_item_id,
                word=clean_word,
                sequence=item.sequence,
                fallback_text=fallback_text,
                provider=self.provider_name,
                error=error,
                failure_type="storage-fallback"
            )

        return CallerAssetResponse(
            callDeckItemId=item.call_deck_item_id,
            word=clean_word,
            sequence=item.sequence,
            line=line,
            audioUrl=audio_url,
            storageKey=storage_key,
            status="ready",
            fallbackText=fallback_text,
            provider=self.provider_name,
            providerMetadata={
                "type": "mock",
                "voiceName": voice_name,
                "tone": tone,
                "speechProvider": speech.provider,
                "speech": speech.metadata,
                "storage": storage_metadata
            }
        )

    def _voice_name(self, voice_name: str | None) -> str:
        clean_voice_name = " ".join((voice_name or "").split())
        return clean_voice_name or self.default_voice_name

    def _tone(self, tone: str | None) -> str:
        clean_tone = " ".join((tone or "").split())
        return clean_tone or "neutral"


class MockCallerAssetProvider:
    provider_name = "mock-python-ai-service"
    default_voice_name = "mock-voice"

    def __init__(
        self,
        speech_provider: SpeechProvider | None = None,
        storage_provider: AudioStorageProvider | None = None
    ):
        self.speech_provider = speech_provider or default_speech_provider()
        self.storage_provider = storage_provider or default_audio_storage_provider()

    def generate_caller_asset(
        self,
        request: CallerAssetRequest
    ) -> CallerAssetResponse:
        clean_word = " ".join(request.word.split())
        tone = self._tone(request.tone)
        voice_name = self._voice_name(request.voice_name)
        fallback_text = fallback_caller_text(clean_word)
        line = build_caller_line(clean_word, tone)
        try:
            speech = self.speech_provider.synthesize(SpeechRequest(
                text=line,
                voice_name=voice_name,
                asset_id=request.call_deck_item_id,
                sequence=request.sequence
            ))
        except Exception as error:
            return failed_caller_asset(
                call_deck_item_id=request.call_deck_item_id,
                word=clean_word,
                sequence=request.sequence,
                fallback_text=fallback_text,
                provider=self.provider_name,
                error=error
            )

        try:
            audio_url, storage_key, storage_metadata = resolve_audio_output(
                speech=speech,
                storage_provider=self.storage_provider
            )
        except Exception as error:
            return failed_caller_asset(
                call_deck_item_id=request.call_deck_item_id,
                word=clean_word,
                sequence=request.sequence,
                fallback_text=fallback_text,
                provider=self.provider_name,
                error=error,
                failure_type="storage-fallback"
            )

        return CallerAssetResponse(
            callDeckItemId=request.call_deck_item_id,
            word=clean_word,
            sequence=request.sequence,
            line=line,
            audioUrl=audio_url,
            storageKey=storage_key,
            status="ready",
            fallbackText=fallback_text,
            provider=self.provider_name,
            providerMetadata={
                "type": "mock",
                "voiceName": voice_name,
                "tone": tone,
                "speechProvider": speech.provider,
                "speech": speech.metadata,
                "storage": storage_metadata
            }
        )

    def _voice_name(self, voice_name: str | None) -> str:
        clean_voice_name = " ".join((voice_name or "").split())
        return clean_voice_name or self.default_voice_name

    def _tone(self, tone: str | None) -> str:
        clean_tone = " ".join((tone or "").split())
        return clean_tone or "neutral"


def fallback_caller_text(word: str) -> str:
    clean_word = " ".join(word.split())
    return f"Next word is {clean_word}."


def build_caller_line(word: str, tone: str | None) -> str:
    clean_word = " ".join(word.split())
    clean_tone = " ".join((tone or "").split()).lower()
    lines = {
        "fun": f"{clean_word} is up. Mark it if you have it!",
        "energetic": f"{clean_word}! Big energy, check your card!",
        "professional": f"{clean_word} is the call. Please check your card.",
        "calm": f"{clean_word} is next. Take a quick look.",
        "neutral": f"{clean_word} is next. Check your card.",
    }
    line = lines.get(clean_tone, f"{clean_word} is next. Check your card.")
    return safe_caller_line(line=line, word=clean_word)


def safe_caller_line(line: str, word: str) -> str:
    clean_line = " ".join((line or "").split())
    clean_word = " ".join(word.split())
    lowered_line = clean_line.lower()
    if not clean_line or any(
        phrase in lowered_line for phrase in BANNED_AI_DECISION_PHRASES
    ):
        return fallback_caller_text(clean_word)
    return clean_line


def sanitize_caller_asset_response(
    response: CallerAssetResponse
) -> CallerAssetResponse:
    original_line = " ".join((response.line or "").split())
    response.line = safe_caller_line(line=response.line, word=response.word)
    if not response.fallback_text:
        response.fallback_text = fallback_caller_text(response.word)
    if response.status == "ready" and response.line != original_line:
        response.status = "fallback"
    return response


def sanitize_caller_assets_bulk_response(
    response: CallerAssetsBulkResponse
) -> CallerAssetsBulkResponse:
    response.assets = [
        sanitize_caller_asset_response(asset)
        for asset in response.assets
    ]
    return response


def sanitize_theme_response(
    response: ThemeResponse,
    allowed_assets: list[str]
) -> ThemeResponse:
    safe_allowed_assets = {
        asset.strip()
        for asset in allowed_assets
        if safe_theme_token(asset)
    }
    response.icons = [
        asset
        for asset in response.icons
        if safe_theme_token(asset) and asset in safe_allowed_assets
    ]
    response.decorations = [
        asset
        for asset in response.decorations
        if safe_theme_token(asset) and asset in safe_allowed_assets
    ]
    response.palette = {
        key: value
        for key, value in response.palette.items()
        if safe_theme_token(key) and safe_theme_value(value)
    }
    response.name = safe_theme_text(response.name, "Mock Bingo Theme")
    response.summary = safe_theme_text(
        response.summary,
        "A safe structured theme using only allowed tokens."
    )
    response.motion = safe_theme_text(response.motion, "subtle")
    response.caller_tone = safe_theme_text(response.caller_tone, "friendly host")
    response.accessibility["contrastChecked"] = True
    response.accessibility["avoidColorOnlyMeaning"] = True
    return response


def safe_theme_token(value: str) -> bool:
    clean_value = value.strip()
    lowered = clean_value.lower()
    return (
        bool(clean_value)
        and "/" not in clean_value
        and "\\" not in clean_value
        and ":" not in clean_value
        and "<" not in clean_value
        and ">" not in clean_value
        and "url(" not in lowered
        and "javascript" not in lowered
        and "script" not in lowered
    )


def safe_theme_value(value: str) -> bool:
    lowered = value.lower()
    return (
        "url(" not in lowered
        and "javascript" not in lowered
        and "<script" not in lowered
        and "http://" not in lowered
        and "https://" not in lowered
        and "{" not in value
        and "}" not in value
    )


def safe_theme_text(value: str, fallback: str) -> str:
    clean_value = " ".join((value or "").split())
    if clean_value and safe_theme_value(clean_value):
        return clean_value
    return fallback


def resolve_audio_output(
    speech,
    storage_provider: AudioStorageProvider
) -> tuple[str | None, str | None, dict[str, object]]:
    if (speech.audio_url or speech.storage_key) and not speech.audio_bytes:
        return speech.audio_url, speech.storage_key, {
            "provider": None,
            "stored": False
        }

    if not speech.audio_bytes:
        return None, speech.storage_key, {
            "provider": None,
            "stored": False
        }

    storage = storage_provider.store(AudioStorageRequest(
        audio_bytes=speech.audio_bytes,
        storage_key=speech.storage_key or "caller-assets/audio.mp3",
        content_type=speech.content_type
    ))
    return storage.audio_url, storage.storage_key, {
        "provider": storage.provider,
        "stored": True,
        "metadata": storage.metadata
    }


def fallback_caller_asset_response(
    request: CallerAssetRequest,
    error: Exception
) -> CallerAssetResponse:
    clean_word = " ".join(request.word.split())
    fallback_text = fallback_caller_text(clean_word)
    return CallerAssetResponse(
        callDeckItemId=request.call_deck_item_id,
        word=clean_word,
        sequence=request.sequence,
        line=fallback_text,
        status="failed",
        fallbackText=fallback_text,
        provider="local-fallback",
        providerMetadata={
            "type": "fallback",
            "error": str(error)
        }
    )


def failed_caller_asset(
    call_deck_item_id: str,
    word: str,
    sequence: int,
    fallback_text: str,
    provider: str,
    error: Exception,
    failure_type: str = "speech-fallback"
) -> CallerAssetResponse:
    return CallerAssetResponse(
        callDeckItemId=call_deck_item_id,
        word=word,
        sequence=sequence,
        line=fallback_text,
        status="failed",
        fallbackText=fallback_text,
        provider=provider,
        providerMetadata={
            "type": failure_type,
            "error": str(error)
        }
    )


class AIGenerationService:
    def __init__(
        self,
        game_prep_provider: GamePrepProvider | None = None,
        caller_assets_bulk_provider: CallerAssetsBulkProvider | None = None,
        caller_asset_provider: CallerAssetProvider | None = None,
        theme_provider: ThemeProvider | None = None
    ):
        self.game_prep_provider = game_prep_provider or MockGamePrepProvider()
        self.caller_assets_bulk_provider = (
            caller_assets_bulk_provider or MockCallerAssetsBulkProvider()
        )
        self.caller_asset_provider = caller_asset_provider or MockCallerAssetProvider()
        self.theme_provider = theme_provider or MockThemeProvider()

    def generate_game_prep(self, request: GamePrepRequest) -> GamePrepResponse:
        return self.game_prep_provider.generate_game_prep(request)

    def generate_caller_assets_bulk(
        self,
        request: CallerAssetsBulkRequest
    ) -> CallerAssetsBulkResponse:
        response = self.caller_assets_bulk_provider.generate_caller_assets_bulk(request)
        return sanitize_caller_assets_bulk_response(response)

    def generate_caller_asset(
        self,
        request: CallerAssetRequest
    ) -> CallerAssetResponse:
        try:
            response = self.caller_asset_provider.generate_caller_asset(request)
            return sanitize_caller_asset_response(response)
        except Exception as error:
            return fallback_caller_asset_response(request, error)

    def generate_theme(self, request: ThemeRequest) -> ThemeResponse:
        return sanitize_theme_response(
            self.theme_provider.generate_theme(request),
            request.allowed_assets
        )
