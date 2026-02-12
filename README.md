# HackersEra AI SDK

Go SDK for the [HackersEra AI](https://hub.docker.com/r/hackerseravsoc/hackersera-ai-model-provider) API.

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
    client := sdk.NewClient("https://your-api-endpoint.com", "your-api-key")

    resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
        Model:    sdk.ModelDefault,
        Messages: []sdk.Message{{Role: "user", Content: "Hello!"}},
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

## Available Models

| Model | Constant | Use Case |
|---|---|---|
| `hackersera-ai` | `sdk.ModelDefault` | General purpose (default) |
| `hackersera-ai-pro` | `sdk.ModelPro` | Complex reasoning |
| `hackersera-ai-lite` | `sdk.ModelLite` | Fast, lightweight tasks |
| `hackersera-ai-embedding` | `sdk.ModelEmbedding` | Text embeddings |

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
    Model: sdk.ModelDefault,
    Messages: []sdk.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "Explain Go interfaces"},
    },
})

fmt.Println(resp.Choices[0].Message.Content)
fmt.Printf("Tokens: %d\n", resp.Usage.TotalTokens)
```

With optional parameters:

```go
resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
    Model: sdk.ModelPro,
    Messages: []sdk.Message{
        {Role: "user", Content: "Write a creative story"},
    },
    MaxTokens:   sdk.IntPtr(2000),
    Temperature: sdk.Float64Ptr(0.9),
    TopP:        sdk.Float64Ptr(0.95),
})
```

### Streaming

```go
chunks, errs := client.ChatCompletionStream(ctx, sdk.ChatRequest{
    Model:    sdk.ModelDefault,
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
model, err := client.GetModel(ctx, sdk.ModelDefault)
```

### Embeddings

```go
emb, err := client.CreateEmbedding(ctx, sdk.EmbeddingRequest{
    Input: "Hello world",
    Model: sdk.ModelEmbedding,
})
```

### Health Check

```go
health, err := client.Health(ctx)
fmt.Printf("Status: %s, Version: %s\n", health.Status, health.Version)
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

## Integrations

### OpenCode

HackersEra AI works as a custom provider in [OpenCode](https://opencode.ai). Add this to your `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "hackersera": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "HackersEra AI",
      "options": {
        "baseURL": "https://your-api-endpoint.com/v1",
        "apiKey": "{env:HACKERSERA_API_KEY}"
      },
      "models": {
        "hackersera-ai": {
          "name": "HackersEra AI",
          "limit": {
            "context": 200000,
            "output": 8192
          }
        },
        "hackersera-ai-pro": {
          "name": "HackersEra AI Pro",
          "limit": {
            "context": 200000,
            "output": 8192
          }
        },
        "hackersera-ai-lite": {
          "name": "HackersEra AI Lite",
          "limit": {
            "context": 200000,
            "output": 8192
          }
        }
      }
    }
  }
}
```

Set your API key:

```bash
export HACKERSERA_API_KEY=your-api-key
```

Or use the `/connect` command in OpenCode, select **Other**, enter `hackersera` as the provider ID, and paste your API key. Then run `/models` to see HackersEra AI models in the selection list.

### Any OpenAI-Compatible Client

HackersEra AI is fully OpenAI-compatible. Use it with any client that supports custom OpenAI endpoints:

**Python (openai)**
```python
from openai import OpenAI

client = OpenAI(
    base_url="https://your-api-endpoint.com/v1",
    api_key="your-api-key",
)

response = client.chat.completions.create(
    model="hackersera-ai",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(response.choices[0].message.content)
```

**curl**
```bash
curl https://your-api-endpoint.com/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "hackersera-ai",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## License

MIT
