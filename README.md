# StoryCode

StoryCode is a local-first CLI that analyzes a Python repository, drafts
interactive stories from FastAPI entry points, and serves them on `127.0.0.1`.

It ships as a single Go binary. Runtime does not require Docker, Node.js,
Python, or an external database. Analysis never executes code from the
target repository.

## Install

```bash
go run ./cmd/storycode --help
```

## First commands

```bash
storycode init
storycode status
```

`index`, `discover`, `serve`, `story`, and `verify` exist as CLI stubs.

## Current limitations

- Story discovery is not implemented.
- No local HTTP API or web UI yet.
- The FastAPI fixture is static analysis input, not a live LLM or RAG service.

## License

MIT. See [LICENSE](LICENSE).
