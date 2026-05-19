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


def validate_called_number(called_number: str) -> bool:
    """
    Checks whether a Bingo call is in the correct format and range.

    This only checks format/range.
    It does NOT confirm whether the number was officially called.
    Official game state must be handled by the Go backend.
    """

    if not called_number or len(called_number) < 2:
        return False

    letter = called_number[0].upper()
    number_part = called_number[1:]

    if letter not in ["B", "I", "N", "G", "O"]:
        return False

    if not number_part.isdigit():
        return False

    number = int(number_part)

    valid_ranges = {
        "B": range(1, 16),
        "I": range(16, 31),
        "N": range(31, 46),
        "G": range(46, 61),
        "O": range(61, 76),
    }

    return number in valid_ranges[letter]