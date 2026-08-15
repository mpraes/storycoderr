# fastapi-rag-demo

Static FastAPI fixture used by StoryCode as an analysis source. StoryCode
reads these files; it must never execute Python, tests, hooks, or imports
from this directory.

## Expected flow

```text
POST /v1/chat
→ create_chat()
→ RetrievalService.retrieve()
→ VectorStore.search()
→ GenerationService.generate()
→ ChatResponse
```

`create_chat` is an async `@router.post` handler on prefix `/v1`. Retrieval
calls the vector store. Generation turns snippets into a `ChatResponse`.

When `VectorStore.search` finds nothing, `RetrievalService.retrieve` returns
`[]` and `GenerationService.generate` answers that no context was found.

## Layout

- `app/api/chat.py` — route and handler
- `app/services/retrieval.py` — `RetrievalService.retrieve`
- `app/repositories/vector_store.py` — `VectorStore.search`
- `app/services/generation.py` — `GenerationService.generate`
- `app/models/chat.py` — `ChatRequest` / `ChatResponse`
- `tests/test_chat.py` — tests named after `POST /v1/chat`
