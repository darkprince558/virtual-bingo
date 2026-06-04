from fastapi import APIRouter

from app.models.theme import BingoTheme
from app.services.theme_service import ThemeService


router = APIRouter()

service = ThemeService()


@router.post(
    "/generate-theme/{topic}",
    response_model=BingoTheme
)
async def generate_theme(
    topic: str
):

    return await service.generate_theme(
        topic
    )
