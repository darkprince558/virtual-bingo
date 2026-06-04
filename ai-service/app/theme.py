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
