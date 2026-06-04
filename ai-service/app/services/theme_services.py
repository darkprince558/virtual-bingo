from pydantic import BaseModel


class ThemePalette(BaseModel):
    primary: str
    secondary: str
    accent: str


class ThemeMusic(BaseModel):
    genre: str
    mood: str
    description: str


class BingoTheme(BaseModel):
    name: str
    icon: str

    palette: ThemePalette

    music: ThemeMusic

    voice_recommendation: str

    phrases: list[str]

class ThemeService:

    async def generate_theme(
        self,
        topic: str
    ) -> BingoTheme:

        response = await client.chat.completions.create(
            model="gpt-4.1-mini",
            response_format={"type": "json_object"},
            messages=[
                {
                    "role": "system",
                    "content": SYSTEM_PROMPT
                },
                {
                    "role": "user",
                    "content": topic
                }
            ],
            temperature=0.9
        )

        theme = json.loads(
            response.choices[0].message.content
        )

        return BingoTheme(**theme)

def build_call_prompt(theme: str, number: int, voice_style: str):

    return f"""
Theme: {theme}
Voice style: {voice_style}

Generate a bingo call for number {number}.

Rules:
- 1 short sentence
- match voice style tone
- family friendly
"""
