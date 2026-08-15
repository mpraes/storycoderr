from app.repositories.vector_store import VectorStore


class RetrievalService:
    def __init__(self, vector_store: VectorStore) -> None:
        self._vector_store = vector_store

    async def retrieve(self, query: str) -> list[str]:
        hits = self._vector_store.search(query)
        if not hits:
            return []
        return hits
