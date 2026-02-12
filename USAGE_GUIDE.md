# HackersEra AI SDK - Comprehensive Usage Guide

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [Authentication](#authentication)
4. [Client Configuration](#client-configuration)
5. [Chat Completions](#chat-completions)
6. [Streaming Responses](#streaming-responses)
7. [Model Management](#model-management)
8. [Embeddings](#embeddings)
9. [Error Handling](#error-handling)
10. [Advanced Usage](#advanced-usage)
11. [Best Practices](#best-practices)
12. [Examples](#examples)

---

## Installation

Install the SDK using Go modules:

```bash
go get github.com/hackersera-dev-team/hackersera-ai-sdk
```

Import in your Go code:

```go
import sdk "github.com/hackersera-dev-team/hackersera-ai-sdk"
```

---

## Quick Start

Here's a minimal example to get started:

```go
package main

import (
    "context"
    "fmt"
    "log"

    sdk "github.com/hackersera-dev-team/hackersera-ai-sdk"
)

func main() {
    // Create a new client
    client := sdk.NewClient(
        "http://hackersera-ai.cloudjiffy.net",
        "your-api-key",
    )

    // Send a chat completion request
    resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
        Model: sdk.ModelDefault,
        Messages: []sdk.Message{
            {Role: "user", Content: "Hello, AI!"},
        },
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

---

## Authentication

### Using API Keys

The SDK requires an API key for authentication. Set it when creating the client:

```go
client := sdk.NewClient("http://hackersera-ai.cloudjiffy.net", "your-api-key")
```

### Using Environment Variables (Recommended)

Store your API key in an environment variable for security:

```bash
export HACKERSERA_API_KEY=your-api-key
```

Then use it in your code:

```go
import "os"

apiKey := os.Getenv("HACKERSERA_API_KEY")
if apiKey == "" {
    log.Fatal("HACKERSERA_API_KEY environment variable is required")
}

client := sdk.NewClient("http://hackersera-ai.cloudjiffy.net", apiKey)
```

### Using .env Files

Create a `.env` file:

```bash
HACKERSERA_API_KEY=your-api-key
```

Load it using a package like `godotenv`:

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file
    godotenv.Load()

    apiKey := os.Getenv("HACKERSERA_API_KEY")
    client := sdk.NewClient("http://hackersera-ai.cloudjiffy.net", apiKey)
}
```

---

## Client Configuration

### Basic Client

```go
client := sdk.NewClient("http://hackersera-ai.cloudjiffy.net", apiKey)
```

### Custom HTTP Client

Configure timeout, retry logic, or custom transport:

```go
import (
    "net/http"
    "time"
)

customHTTPClient := &http.Client{
    Timeout: 10 * time.Minute,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

client := sdk.NewClient("http://hackersera-ai.cloudjiffy.net", apiKey).
    WithHTTPClient(customHTTPClient)
```

### Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.ChatCompletion(ctx, req)
```

---

## Chat Completions

### Simple Chat

```go
resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
    Model: sdk.ModelDefault,
    Messages: []sdk.Message{
        {Role: "user", Content: "What is Go?"},
    },
})

if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message.Content)
```

### Multi-turn Conversation

```go
messages := []sdk.Message{
    {Role: "system", Content: "You are a helpful programming assistant."},
    {Role: "user", Content: "What is a goroutine?"},
    {Role: "assistant", Content: "A goroutine is a lightweight thread..."},
    {Role: "user", Content: "How do I create one?"},
}

resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
    Model:    sdk.ModelDefault,
    Messages: messages,
})
```

### Using Different Models

```go
// Default model (general purpose)
resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
    Model:    sdk.ModelDefault, // hackersera-ai
    Messages: messages,
})

// Pro model (complex reasoning)
resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
    Model:    sdk.ModelPro, // hackersera-ai-pro
    Messages: messages,
})

// Lite model (fast, lightweight)
resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
    Model:    sdk.ModelLite, // hackersera-ai-lite
    Messages: messages,
})
```

### Advanced Parameters

```go
resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
    Model: sdk.ModelPro,
    Messages: []sdk.Message{
        {Role: "user", Content: "Write a creative story"},
    },
    MaxTokens:   sdk.IntPtr(2000),        // Limit response length
    Temperature: sdk.Float64Ptr(0.9),     // Higher = more creative
    TopP:        sdk.Float64Ptr(0.95),    // Nucleus sampling
    Stop:        []string{"\n\n", "END"}, // Stop sequences
})
```

### Accessing Response Metadata

```go
resp, err := client.ChatCompletion(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Model: %s\n", resp.Model)
fmt.Printf("Finish Reason: %s\n", resp.Choices[0].FinishReason)
fmt.Printf("Prompt Tokens: %d\n", resp.Usage.PromptTokens)
fmt.Printf("Completion Tokens: %d\n", resp.Usage.CompletionTokens)
fmt.Printf("Total Tokens: %d\n", resp.Usage.TotalTokens)
fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
```

---

## Streaming Responses

### Basic Streaming

```go
chunks, errs := client.ChatCompletionStream(context.Background(), sdk.ChatRequest{
    Model: sdk.ModelDefault,
    Messages: []sdk.Message{
        {Role: "user", Content: "Write a poem about Go"},
    },
})

for {
    select {
    case chunk, ok := <-chunks:
        if !ok {
            // Stream finished
            return
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

### Streaming with Progress Tracking

```go
var fullResponse strings.Builder
tokenCount := 0

chunks, errs := client.ChatCompletionStream(ctx, req)

for {
    select {
    case chunk, ok := <-chunks:
        if !ok {
            fmt.Printf("\n\nTotal tokens received: %d\n", tokenCount)
            fmt.Printf("Full response:\n%s\n", fullResponse.String())
            return
        }

        content := chunk.Choices[0].Delta.Content
        fullResponse.WriteString(content)
        tokenCount++

        // Print to console in real-time
        fmt.Print(content)

    case err, ok := <-errs:
        if ok && err != nil {
            log.Printf("Streaming error: %v\n", err)
        }
        return
    }
}
```

### Streaming with Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

chunks, errs := client.ChatCompletionStream(ctx, req)

// Cancel after 10 seconds or on user interrupt
go func() {
    time.Sleep(10 * time.Second)
    cancel()
}()

for {
    select {
    case chunk, ok := <-chunks:
        if !ok {
            return
        }
        fmt.Print(chunk.Choices[0].Delta.Content)

    case <-ctx.Done():
        fmt.Println("\nStream cancelled")
        return

    case err := <-errs:
        if err != nil {
            log.Fatal(err)
        }
        return
    }
}
```

---

## Model Management

### List All Models

```go
models, err := client.ListModels(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Available models:\n")
for _, model := range models.Data {
    fmt.Printf("- %s (owned by: %s)\n", model.ID, model.OwnedBy)
    fmt.Printf("  Created: %s\n", time.Unix(model.Created, 0))
}
```

### Get Model Details

```go
model, err := client.GetModel(context.Background(), sdk.ModelDefault)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Model ID: %s\n", model.ID)
fmt.Printf("Object Type: %s\n", model.Object)
fmt.Printf("Owned By: %s\n", model.OwnedBy)
fmt.Printf("Created: %s\n", time.Unix(model.Created, 0))
```

### Check Model Availability

```go
func isModelAvailable(client *sdk.Client, modelID string) bool {
    _, err := client.GetModel(context.Background(), modelID)
    return err == nil
}

if isModelAvailable(client, sdk.ModelPro) {
    fmt.Println("Pro model is available")
}
```

---

## Embeddings

### Create Embeddings

```go
resp, err := client.CreateEmbedding(context.Background(), sdk.EmbeddingRequest{
    Input: "Hello, world!",
    Model: sdk.ModelEmbedding,
})

if err != nil {
    log.Fatal(err)
}

embedding := resp.Data[0].Embedding
fmt.Printf("Embedding dimensions: %d\n", len(embedding))
fmt.Printf("First 5 values: %v\n", embedding[:5])
```

### Batch Embeddings

```go
texts := []string{
    "The quick brown fox",
    "jumps over the lazy dog",
    "Go is a compiled language",
}

for i, text := range texts {
    resp, err := client.CreateEmbedding(ctx, sdk.EmbeddingRequest{
        Input: text,
        Model: sdk.ModelEmbedding,
    })

    if err != nil {
        log.Printf("Error embedding text %d: %v\n", i, err)
        continue
    }

    fmt.Printf("Text %d: %d dimensions\n", i, len(resp.Data[0].Embedding))
}
```

### Cosine Similarity

```go
import "math"

func cosineSimilarity(a, b []float64) float64 {
    var dotProduct, normA, normB float64

    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }

    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Get embeddings for two texts
emb1, _ := client.CreateEmbedding(ctx, sdk.EmbeddingRequest{
    Input: "artificial intelligence",
    Model: sdk.ModelEmbedding,
})

emb2, _ := client.CreateEmbedding(ctx, sdk.EmbeddingRequest{
    Input: "machine learning",
    Model: sdk.ModelEmbedding,
})

similarity := cosineSimilarity(
    emb1.Data[0].Embedding,
    emb2.Data[0].Embedding,
)

fmt.Printf("Similarity: %.4f\n", similarity)
```

---

## Error Handling

### Basic Error Handling

```go
resp, err := client.ChatCompletion(ctx, req)
if err != nil {
    log.Fatalf("API error: %v", err)
}
```

### Detailed Error Information

```go
resp, err := client.ChatCompletion(ctx, req)
if err != nil {
    if apiErr, ok := err.(*sdk.APIError); ok {
        fmt.Printf("Status Code: %d\n", apiErr.StatusCode)
        fmt.Printf("Error Type: %s\n", apiErr.ErrorBody.Error.Type)
        fmt.Printf("Error Message: %s\n", apiErr.ErrorBody.Error.Message)

        // Handle specific error types
        switch apiErr.ErrorBody.Error.Type {
        case "invalid_api_key":
            log.Fatal("Invalid API key provided")
        case "rate_limit_exceeded":
            log.Fatal("Rate limit exceeded, please retry later")
        case "invalid_request_error":
            log.Fatal("Invalid request parameters")
        default:
            log.Fatalf("API error: %s", apiErr.ErrorBody.Error.Message)
        }
    } else {
        log.Fatalf("Unexpected error: %v", err)
    }
}
```

### Retry Logic

```go
func chatWithRetry(client *sdk.Client, req sdk.ChatRequest, maxRetries int) (*sdk.ChatResponse, error) {
    var resp *sdk.ChatResponse
    var err error

    for i := 0; i < maxRetries; i++ {
        resp, err = client.ChatCompletion(context.Background(), req)

        if err == nil {
            return resp, nil
        }

        // Check if error is retryable
        if apiErr, ok := err.(*sdk.APIError); ok {
            if apiErr.StatusCode >= 500 {
                // Server error, retry with backoff
                time.Sleep(time.Duration(i+1) * time.Second)
                continue
            }
        }

        // Non-retryable error
        return nil, err
    }

    return nil, fmt.Errorf("max retries exceeded: %w", err)
}
```

---

## Advanced Usage

### Concurrent Requests

```go
import "sync"

func processConcurrent(client *sdk.Client, prompts []string) {
    var wg sync.WaitGroup
    results := make(chan string, len(prompts))

    for _, prompt := range prompts {
        wg.Add(1)
        go func(p string) {
            defer wg.Done()

            resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
                Model:    sdk.ModelDefault,
                Messages: []sdk.Message{{Role: "user", Content: p}},
            })

            if err != nil {
                results <- fmt.Sprintf("Error: %v", err)
                return
            }

            results <- resp.Choices[0].Message.Content
        }(prompt)
    }

    wg.Wait()
    close(results)

    for result := range results {
        fmt.Println(result)
    }
}
```

### Rate Limiting

```go
import "golang.org/x/time/rate"

type RateLimitedClient struct {
    client  *sdk.Client
    limiter *rate.Limiter
}

func NewRateLimitedClient(client *sdk.Client, rps float64) *RateLimitedClient {
    return &RateLimitedClient{
        client:  client,
        limiter: rate.NewLimiter(rate.Limit(rps), 1),
    }
}

func (c *RateLimitedClient) ChatCompletion(ctx context.Context, req sdk.ChatRequest) (*sdk.ChatResponse, error) {
    if err := c.limiter.Wait(ctx); err != nil {
        return nil, err
    }
    return c.client.ChatCompletion(ctx, req)
}
```

### Response Caching

```go
import "sync"

type CachedClient struct {
    client *sdk.Client
    cache  map[string]*sdk.ChatResponse
    mu     sync.RWMutex
}

func NewCachedClient(client *sdk.Client) *CachedClient {
    return &CachedClient{
        client: client,
        cache:  make(map[string]*sdk.ChatResponse),
    }
}

func (c *CachedClient) ChatCompletion(ctx context.Context, req sdk.ChatRequest) (*sdk.ChatResponse, error) {
    key := fmt.Sprintf("%v", req.Messages)

    // Check cache
    c.mu.RLock()
    if cached, ok := c.cache[key]; ok {
        c.mu.RUnlock()
        return cached, nil
    }
    c.mu.RUnlock()

    // Call API
    resp, err := c.client.ChatCompletion(ctx, req)
    if err != nil {
        return nil, err
    }

    // Store in cache
    c.mu.Lock()
    c.cache[key] = resp
    c.mu.Unlock()

    return resp, nil
}
```

### Health Monitoring

```go
func monitorHealth(client *sdk.Client, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for range ticker.C {
        health, err := client.Health(context.Background())
        if err != nil {
            log.Printf("Health check failed: %v\n", err)
            continue
        }

        if health.Status != "ok" {
            log.Printf("Warning: API status is %s\n", health.Status)
        } else {
            log.Printf("API healthy (version: %s)\n", health.Version)
        }
    }
}

// Run in background
go monitorHealth(client, 1*time.Minute)
```

---

## Best Practices

### 1. Always Use Context

```go
// Good
ctx := context.Background()
resp, err := client.ChatCompletion(ctx, req)

// Better - with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
resp, err := client.ChatCompletion(ctx, req)
```

### 2. Handle Errors Properly

```go
resp, err := client.ChatCompletion(ctx, req)
if err != nil {
    // Log error with context
    log.Printf("ChatCompletion failed: %v (model: %s)", err, req.Model)
    return err
}
```

### 3. Use Environment Variables for Secrets

```go
// Never hardcode API keys
apiKey := os.Getenv("HACKERSERA_API_KEY")
```

### 4. Choose the Right Model

```go
// Simple tasks - use lite model (faster, cheaper)
resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
    Model:    sdk.ModelLite,
    Messages: []sdk.Message{{Role: "user", Content: "What is 2+2?"}},
})

// Complex reasoning - use pro model
resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
    Model:    sdk.ModelPro,
    Messages: []sdk.Message{{Role: "user", Content: "Explain quantum computing"}},
})
```

### 5. Monitor Token Usage

```go
totalTokens := 0

for _, prompt := range prompts {
    resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
        Model:    sdk.ModelDefault,
        Messages: []sdk.Message{{Role: "user", Content: prompt}},
    })

    if err != nil {
        continue
    }

    totalTokens += resp.Usage.TotalTokens
}

fmt.Printf("Total tokens used: %d\n", totalTokens)
```

### 6. Use Streaming for Long Responses

```go
// For long-form content, use streaming for better UX
chunks, errs := client.ChatCompletionStream(ctx, sdk.ChatRequest{
    Model:    sdk.ModelDefault,
    Messages: []sdk.Message{{Role: "user", Content: "Write a long article"}},
})
```

---

## Examples

### Example 1: CLI Chatbot

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    sdk "github.com/hackersera-dev-team/hackersera-ai-sdk"
)

func main() {
    client := sdk.NewClient(
        "http://hackersera-ai.cloudjiffy.net",
        os.Getenv("HACKERSERA_API_KEY"),
    )

    messages := []sdk.Message{
        {Role: "system", Content: "You are a helpful assistant."},
    }

    scanner := bufio.NewScanner(os.Stdin)
    fmt.Println("Chatbot (type 'exit' to quit):")

    for {
        fmt.Print("\nYou: ")
        if !scanner.Scan() {
            break
        }

        input := strings.TrimSpace(scanner.Text())
        if input == "exit" {
            break
        }

        messages = append(messages, sdk.Message{
            Role:    "user",
            Content: input,
        })

        resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
            Model:    sdk.ModelDefault,
            Messages: messages,
        })

        if err != nil {
            log.Printf("Error: %v\n", err)
            continue
        }

        assistant := resp.Choices[0].Message.Content
        messages = append(messages, sdk.Message{
            Role:    "assistant",
            Content: assistant,
        })

        fmt.Printf("\nAssistant: %s\n", assistant)
    }
}
```

### Example 2: Document Summarizer

```go
func summarizeDocument(client *sdk.Client, document string) (string, error) {
    resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
        Model: sdk.ModelDefault,
        Messages: []sdk.Message{
            {Role: "system", Content: "You are a document summarizer. Provide concise summaries."},
            {Role: "user", Content: fmt.Sprintf("Summarize this document:\n\n%s", document)},
        },
        MaxTokens: sdk.IntPtr(500),
    })

    if err != nil {
        return "", err
    }

    return resp.Choices[0].Message.Content, nil
}
```

### Example 3: Code Generator

```go
func generateCode(client *sdk.Client, description string, language string) (string, error) {
    prompt := fmt.Sprintf("Write %s code for: %s", language, description)

    resp, err := client.ChatCompletion(context.Background(), sdk.ChatRequest{
        Model: sdk.ModelPro,
        Messages: []sdk.Message{
            {Role: "system", Content: "You are an expert programmer. Only output code, no explanations."},
            {Role: "user", Content: prompt},
        },
        Temperature: sdk.Float64Ptr(0.2), // Lower temperature for more deterministic code
    })

    if err != nil {
        return "", err
    }

    return resp.Choices[0].Message.Content, nil
}
```

### Example 4: Semantic Search

```go
func semanticSearch(client *sdk.Client, query string, documents []string) (string, error) {
    // Get query embedding
    queryEmb, err := client.CreateEmbedding(context.Background(), sdk.EmbeddingRequest{
        Input: query,
        Model: sdk.ModelEmbedding,
    })
    if err != nil {
        return "", err
    }

    // Get document embeddings and find most similar
    var bestDoc string
    var bestSimilarity float64

    for _, doc := range documents {
        docEmb, err := client.CreateEmbedding(context.Background(), sdk.EmbeddingRequest{
            Input: doc,
            Model: sdk.ModelEmbedding,
        })
        if err != nil {
            continue
        }

        similarity := cosineSimilarity(
            queryEmb.Data[0].Embedding,
            docEmb.Data[0].Embedding,
        )

        if similarity > bestSimilarity {
            bestSimilarity = similarity
            bestDoc = doc
        }
    }

    return bestDoc, nil
}
```

---

## Support

For issues, questions, or contributions:
- GitHub: https://github.com/hackersera-dev-team/hackersera-ai-sdk
- Documentation: http://hackersera-ai.cloudjiffy.net

## License

MIT
