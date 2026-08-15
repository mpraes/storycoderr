from app.models.chat import ChatResponse


class GenerationService:
    async def generate(self, query: str, snippets: list[str]) -> ChatResponse:
        if not snippets:
            return ChatResponse(answer=f"No context found for: {query}", context=[])
        return ChatResponse(answer=" ".join(snippets), context=list(snippets))
