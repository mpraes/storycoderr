# Coding standards

These rules apply to StoryCode itself. They do not require Node.js, Python,
or Docker when running the `storycode` binary.

## Formatting

- Go is formatted with `gofmt`. Run `make format`.
- TypeScript and other frontend files are formatted with Prettier. Run `make format`.
- CI rejects unformatted files via `make format-check`.

## Linting

- Go is linted with `go vet`.
- Frontend is linted with ESLint when TypeScript files exist.
- CI runs `make lint` on every pull request.

## Naming and copy

- Entity, field, package, and function names are in English.
- User-facing CLI and UI copy starts in English.
- HTTP JSON contracts use `snake_case` keys.

## Errors

- Public errors are structured values with a stable machine-readable code and a
  human message.
- Messages must include the offending value and the expected shape.
- Never log secrets from the analyzed repository.

## Tests

- Domain logic and HTTP handlers require tests.
- Tests must not execute code, scripts, hooks, or imports from the analyzed
  repository. Use fixtures such as `fixtures/fastapi-rag-demo` as static files.
