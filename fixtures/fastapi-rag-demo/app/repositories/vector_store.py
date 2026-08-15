class VectorStore:
    def __init__(self) -> None:
        self._chunks = {
            "rag": [
                "Retrieval finds relevant snippets.",
                "Generation writes an answer from those snippets.",
            ],
        }

    def search(self, query: str) -> list[str]:
        needle = query.strip().lower()
        for term, snippets in self._chunks.items():
            if term in needle:
                return list(snippets)
        return []
