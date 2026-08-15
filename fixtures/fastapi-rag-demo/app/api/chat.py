from fastapi import APIRouter

from app.models.chat import ChatRequest, ChatResponse
from app.repositories.vector_store import VectorStore
from app.services.generation import GenerationService
from app.services.retrieval import RetrievalService

router = APIRouter(prefix="/v1")
_retrieval = RetrievalService(VectorStore())
_generation = GenerationService()


@router.post("/chat")
async def create_chat(request: ChatRequest) -> ChatResponse:
    snippets = await _retrieval.retrieve(request.query)
    return await _generation.generate(request.query, snippets)
