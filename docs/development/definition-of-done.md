# Definition of done

A task is done only when all of the following hold:

1. One behavior was implemented and validated before starting the next task.
2. `make format-check` passes.
3. `make lint` passes.
4. `make test` passes, including tests that read `fixtures/fastapi-rag-demo`
   without executing Python.
5. `make build` produces a local binary.
6. New domain logic and HTTP handlers have tests.
7. The change does not require Docker, Node.js, Python, or an external
   database to run `storycode`.
8. The analyzed repository is treated as untrusted content. Secrets are never
   written to logs, the browser, or exports.
9. No story discovery, frontend UI, or HTTP API was added unless the task
   explicitly asks for it.
