from fastapi import APIRouter

router = APIRouter(prefix="/v1")


@router.post("/chat")
async def create_chat() -> dict[str, str]:
    return {"status": "ok"}
