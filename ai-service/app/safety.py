BANNED_AI_DECISION_PHRASES = [
    "winner confirmed",
    "player has won",
    "valid bingo",
    "invalid bingo",
    "game result",
    "the game is over",
    "official winner",
    "confirmed winner",
]


def ensure_non_critical_commentary(text: str) -> str:
    """
    Ensures the AI service only provides narration/commentary.

    The AI service should NOT control:
    - Winner validation
    - Player card state
    - Final game outcomes
    - Official Bingo results

    Those responsibilities belong to the Go game backend,
    which remains the source of truth.
    """

    lowered_text = text.lower()

    for phrase in BANNED_AI_DECISION_PHRASES:
        if phrase in lowered_text:
            return (
                "I can provide commentary, but official game results, "
                "winner validation, card state, and final decisions are "
                "handled by the Bingo game system."
            )

    return text
