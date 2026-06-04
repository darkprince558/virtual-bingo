import json

from openai import AsyncOpenAI

from app.models.theme import BingoTheme


client = AsyncOpenAI()


SYSTEM_PROMPT = """
You are a theme generator for an AI Bingo game.

Input can be:

- holidays
- festivals
- books
- TV shows
- sports
- musicians
- movies
- historical periods
- academic subjects
- fandoms

Generate:

1. name
2. icon emoji
3. color palette
4. music genre
5. music mood
6. music description
7. recommended voice style
8. 10 bingo call phrases

Rules:

- family friendly
- fun
- short phrases
- return JSON only

Schema:

{
"name":"",
"icon":"",
"palette":{
"primary":"",
"secondary":"",
"accent":""
},
"music":{
"genre":"",
"mood":"",
"description":""
},
"voice_recommendation":"",
"phrases":[]
}
"""

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
