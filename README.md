# hackersera-ai-model-sdk

Go SDK for the [hackersera-ai-model-provider](https://hub.docker.com/r/hackerseravsoc/hackersera-ai-model-provider) API.

## Install

```bash
go get github.com/hackersera-dev-team/hackersera-ai-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    sdk "github.com/hackersera-dev-team/hackersera-ai-sdk"
)

func main() {
    client := sdk.NewClient("http://localhost:8080", "your-api-key")

    resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
        Model:    "sonnet",
        Messages: []sdk.Message{{Role: "user", Content: "Hello!"}},
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## Available Models

| Model ID | Backend | Use Case |
|---|---|---|
| `sonnet` | GLM-4.7 | General purpose (default) |
| `opus` | GLM-4.7 | Complex reasoning |
| `haiku` | GLM-4.5-Air | Fast, lightweight tasks |
| `gpt-4` / `gpt-4o` | GLM-4.7 | OpenAI-compatible alias |
| `gpt-4o-mini` / `gpt-3.5-turbo` | GLM-4.5-Air | OpenAI-compatible alias |

## API Reference

### Create Client

```go
client := sdk.NewClient(baseURL, apiKey)

// With custom HTTP client
client = sdk.NewClient(baseURL, apiKey).WithHTTPClient(&http.Client{
    Timeout: 10 * time.Minute,
})
```

### Chat Completion

```go
resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
    Model: "sonnet",
    Messages: []sdk.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Explain Go interfaces"},
    },
})

fmt.Println(resp.Choices[0].Message.Content)
fmt.Printf("Tokens: %d\n", resp.Usage.TotalTokens)
```

### Streaming

```go
chunks, errs := client.ChatCompletionStream(ctx, sdk.ChatRequest{
    Model:    "sonnet",
    Messages: []sdk.Message{{Role: "user", Content: "Write a poem"}},
})

for {
    select {
    case chunk, ok := <-chunks:
        if !ok {
            return // stream ended
        }
        fmt.Print(chunk.Choices[0].Delta.Content)
    case err, ok := <-errs:
        if ok && err != nil {
            log.Fatal(err)
        }
        return
    }
}
```

### List Models

```go
models, err := client.ListModels(ctx)
for _, m := range models.Data {
    fmt.Printf("%s (%s)\n", m.ID, m.OwnedBy)
}
```

### Get Model

```go
model, err := client.GetModel(ctx, "sonnet")
```

### Embeddings

```go
emb, err := client.CreateEmbedding(ctx, sdk.EmbeddingRequest{
    Input: "Hello world",
    Model: "text-embedding-ada-002",
})
```

### Health Check

```go
health, err := client.Health(ctx)
fmt.Printf("Status: %s\n", health.Status)
```

### Error Handling

```go
resp, err := client.ChatCompletion(ctx, req)
if err != nil {
    if apiErr, ok := err.(*sdk.APIError); ok {
        fmt.Printf("API error %d: %s (type: %s)\n",
            apiErr.StatusCode,
            apiErr.ErrorBody.Error.Message,
            apiErr.ErrorBody.Error.Type,
        )
    }
}
```

## Running the Provider

```bash
docker run -p 8080:8080 \
  -e ANTHROPIC_AUTH_TOKEN="your-zai-api-key" \
  -e API_KEY="your-proxy-api-key" \
  hackerseravsoc/hackersera-ai-model-provider
```

## License

MIT
