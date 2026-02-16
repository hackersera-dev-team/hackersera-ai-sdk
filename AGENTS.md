# AGENTS.md — HackersEra AI SDK

## Project Overview

Go SDK client for the HackersEra AI API (OpenAI-compatible, RAG-powered).
Single-package library (`package hackeserasdk`) with zero external dependencies.
Module path: `github.com/hackersera-dev-team/hackersera-ai-sdk`

## Repository Structure

```
client.go          # SDK client — all API methods (chat, models, embeddings, documents, search, usage, health)
types.go           # All request/response types, model constants, error types, helper functions
examples/main.go   # Runnable demo exercising every endpoint
test/              # Deployment integration test (separate go module with `replace` directive)
```

## Build & Validation Commands

```bash
# Compile-check the library (no binary produced for library packages)
go build ./...

# Run the Go vet static analyzer
go vet ./...

# Format all Go files (check for drift)
gofmt -l .

# Format all Go files (apply fixes)
gofmt -w .

# Tidy module dependencies
go mod tidy
```

### Running Tests

There are no `_test.go` unit tests yet. When adding them, follow standard Go conventions:

```bash
# Run all tests
go test ./...

# Run a single test by name
go test ./... -run TestChatCompletion

# Run tests in a specific package
go test ./... -v

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Deployment Integration Test

The `test/` directory is a separate Go module. It hits a live API and requires credentials:

```bash
export HACKERSERA_API_KEY=your-api-key
cd test && go run test_deployment.go
```

### Running the Example

```bash
export HACKERSERA_API_KEY=your-api-key
cd examples && go run main.go
```

## Code Style Guidelines

### Package Naming

- Package name is `hackeserasdk` (single word, lowercase, no underscores).
- All exported symbols live in this one package — no sub-packages.

### Imports

- Standard library only — zero third-party dependencies.
- Group imports in a single block; `goimports` ordering (stdlib first, then external).
- Common stdlib imports used: `bufio`, `bytes`, `context`, `encoding/json`, `fmt`, `io`, `net/http`, `strings`, `time`.

### Naming Conventions

- **Exported types**: PascalCase nouns — `ChatRequest`, `ChatResponse`, `APIError`.
- **Exported methods**: PascalCase verbs — `ChatCompletion`, `ListModels`, `CreateEmbedding`.
- **Unexported helpers**: camelCase — `setHeaders`, `parseError`.
- **Constants**: PascalCase with category prefix — `ModelDefault`, `ModelPro`, `ModelLite`.
- **Receiver variable**: single letter `c` for `*Client`.
- **Request/Response pairs**: `FooRequest` / `FooResponse` naming pattern.
- **Helper pointer functions**: `IntPtr`, `Float64Ptr` — short, obvious names.

### Types & Structs

- Every exported type and field has a doc comment.
- Struct fields use JSON tags: `json:"field_name"` (snake_case in JSON).
- Optional fields use `omitempty`: `json:"field,omitempty"`.
- Optional numeric parameters use pointer types (`*int`, `*float64`) to distinguish zero from unset.
- Group related types under section-header comments using box-drawing characters:
  ```go
  // ─── Chat Completions ───────────────────────────────────────────
  ```

### Error Handling

- Every error is wrapped with `fmt.Errorf("action: %w", err)` for context.
- Wrap messages are short lowercase verb phrases: `"marshal request"`, `"create request"`, `"send request"`, `"decode response"`, `"read stream"`.
- API errors use a custom `*APIError` type implementing the `error` interface.
- `parseError` reads the body and attempts JSON decode; falls back to raw body as message.
- Callers type-assert with `err.(*APIError)` to access `StatusCode` and `ErrorBody`.
- Non-OK HTTP status is always checked before decoding the success response.

### Method Pattern

Every API method follows the same structure — keep this consistent when adding new endpoints:

```go
func (c *Client) MethodName(ctx context.Context, req RequestType) (*ResponseType, error) {
    // 1. Marshal request body (for POST/PUT)
    body, err := json.Marshal(req)

    // 2. Create HTTP request with context
    httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/v1/path", reader)

    // 3. Set headers
    c.setHeaders(httpReq)

    // 4. Execute request
    resp, err := c.httpClient.Do(httpReq)
    defer resp.Body.Close()

    // 5. Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, c.parseError(resp)
    }

    // 6. Decode response
    var result ResponseType
    json.NewDecoder(resp.Body).Decode(&result)

    return &result, nil
}
```

### Streaming

- Streaming methods return `(<-chan ChunkType, <-chan error)` — both channels are closed when done.
- Channels are buffered (`make(chan T, 100)`).
- SSE parsing: skip empty lines, strip `"data: "` prefix, stop on `"[DONE]"`.
- Streaming uses a separate `http.Client{}` without timeout.
- Respect `ctx.Done()` in the select loop.

### Comments & Documentation

- Every exported type, function, constant, and method has a `//` doc comment.
- Doc comments start with the symbol name: `// Client is the SDK client for...`
- Section headers use box-drawing: `// ─── Section Name ───────────`
- Usage examples go in doc comments where helpful (see `NewClient`).

### Formatting

- Standard `gofmt` formatting — tabs for indentation.
- No line length limit enforced beyond gofmt defaults.
- Blank line between logical sections within functions.

### HTTP & Headers

- `Content-Type: application/json` on all requests.
- `Authorization: Bearer <key>` when API key is set (omitted if empty).
- Base URL trailing slashes are trimmed in `NewClient`.
- Default HTTP client timeout: `5 * time.Minute`.

### Environment Variables

- `HACKERSERA_API_KEY` — API authentication key.
- `HACKERSERA_BASE_URL` / `HACKERSERA_API_URL` — override base URL (used in examples/tests).
- Never commit `.env` files (listed in `.gitignore`).

## Adding a New API Endpoint

1. Add request/response types to `types.go` under the appropriate section header.
2. Add the client method to `client.go` following the standard method pattern above.
3. Add JSON tags with `snake_case` names and `omitempty` for optional fields.
4. Write a `_test.go` unit test (use `httptest.NewServer` to mock the API).
5. Update `examples/main.go` and `test/test_deployment.go` if the endpoint is user-facing.
6. Run `go vet ./...` and `gofmt -w .` before committing.
