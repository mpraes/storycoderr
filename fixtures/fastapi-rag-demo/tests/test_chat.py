import asyncio

from app.models.chat import ChatRequest
from app.repositories.vector_store import VectorStore
from app.services.generation import GenerationService
from app.services.retrieval import RetrievalService


def _create_chat(query: str):
    request = ChatRequest(query=query)
    retrieval = RetrievalService(VectorStore())
    generation = GenerationService()
    snippets = asyncio.run(retrieval.retrieve(request.query))
    return asyncio.run(generation.generate(request.query, snippets))


def test_post_v1_chat_returns_chat_response():
    response = _create_chat("explain rag")
    assert response.answer
    assert response.context


def test_post_v1_chat_without_context():
    response = _create_chat("unknown topic")
    assert response.context == []
    assert "No context found" in response.answer
