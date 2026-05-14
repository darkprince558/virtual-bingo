from app.safety import ensure_non_critical_commentary


class NarrationService:
    def generate_caller_script(
        self,
        called_number: str,
        tone: str = "fun",
        round_number: int | None = None
    ) -> str:
        letter = called_number[0].upper()
        number = called_number[1:]

        if tone == "fun":
            script = (
                f"Alright everyone, here we go! "
                f"The next number is {letter}-{number}. "
                f"Check those cards!"
            )

        elif tone == "energetic":
            script = (
                f"Get ready, players! "
                f"{letter}-{number}! "
                f"Mark it if you have it!"
            )

        elif tone == "professional":
            script = (
                f"The next Bingo call is {letter}-{number}. "
                f"Please mark it if it appears on your card."
            )

        elif tone == "calm":
            script = (
                f"The next call is {letter}-{number}. "
                f"Take a moment to check your card."
            )

        else:
            script = (
                f"The next number is {letter}-{number}. "
                f"Please check your Bingo card."
            )

        if round_number is not None:
            script += f" This is call number {round_number}."

        return ensure_non_critical_commentary(script)

    def generate_description(
        self,
        called_number: str,
        audience: str = "general"
    ) -> str:
        letter = called_number[0].upper()
        number = called_number[1:]

        description = (
            f"The current Bingo call is {letter}-{number}. "
            f"Players should look under the {letter} column "
            f"for number {number}."
        )

        if audience == "new_players":
            description += (
                " In Bingo, each letter represents a column on the card. "
                "Only check the matching column for that number."
            )

        return ensure_non_critical_commentary(description)

    def generate_host_commentary(
        self,
        context: str,
        tone: str = "professional"
    ) -> str:
        if tone == "fun":
            suggestion = (
                "You could say: Everyone keep your eyes on the board. "
                "Things are starting to heat up!"
            )

        elif tone == "energetic":
            suggestion = (
                "You could say: Great energy so far! Stay focused, "
                "the next call could be the one you need!"
            )

        elif tone == "calm":
            suggestion = (
                "You could say: Take your time checking your cards. "
                "We will continue with the next number shortly."
            )

        else:
            suggestion = (
                "You could say: We are continuing with the next Bingo call. "
                "Please review your cards carefully."
            )

        if context:
            suggestion += f" Context considered: {context}"

        return ensure_non_critical_commentary(suggestion)