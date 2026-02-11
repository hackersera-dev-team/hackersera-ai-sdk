package hackeserasdk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the SDK client for the hackersera-ai-model-provider API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new SDK client.
//
//	client := hackeserasdk.NewClient("http://localhost:8080", "your-api-key")
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// WithHTTPClient sets a custom http.Client for the SDK client.
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	c.httpClient = httpClient
	return c
}

// ─── Chat Completions ───────────────────────────────────────────────────────

// ChatCompletion sends a non-streaming chat completion request.
func (c *Client) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	req.Stream = false

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &chatResp, nil
}

// ChatCompletionStream sends a streaming chat completion request.
// Returns a channel that emits ChatStreamChunk values.
// The channel is closed when the stream ends.
func (c *Client) ChatCompletionStream(ctx context.Context, req ChatRequest) (<-chan ChatStreamChunk, <-chan error) {
	chunks := make(chan ChatStreamChunk, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		req.Stream = true

		body, err := json.Marshal(req)
		if err != nil {
			errs <- fmt.Errorf("marshal request: %w", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
		if err != nil {
			errs <- fmt.Errorf("create request: %w", err)
			return
		}
		c.setHeaders(httpReq)

		// Use a client without timeout for streaming
		streamClient := &http.Client{}
		resp, err := streamClient.Do(httpReq)
		if err != nil {
			errs <- fmt.Errorf("send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errs <- c.parseError(resp)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			// Remove "data: " prefix
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			// End of stream
			if data == "[DONE]" {
				return
			}

			var chunk ChatStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			select {
			case chunks <- chunk:
			case <-ctx.Done():
				return
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("read stream: %w", err)
		}
	}()

	return chunks, errs
}

// ─── Models ─────────────────────────────────────────────────────────────────

// ListModels returns all available models.
func (c *Client) ListModels(ctx context.Context) (*ModelList, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var models ModelList
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &models, nil
}

// GetModel returns a specific model by ID.
func (c *Client) GetModel(ctx context.Context, modelID string) (*Model, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/models/"+modelID, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var model Model
	if err := json.NewDecoder(resp.Body).Decode(&model); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &model, nil
}

// ─── Embeddings ─────────────────────────────────────────────────────────────

// CreateEmbedding creates an embedding for the given input text.
func (c *Client) CreateEmbedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &embResp, nil
}

// ─── Health ─────────────────────────────────────────────────────────────────

// Health checks the health of the API server.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		return nil, c.parseError(resp)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &health, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			ErrorBody: ErrorResponse{
				Error: ErrorDetail{
					Message: string(body),
					Type:    "unknown_error",
				},
			},
		}
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		ErrorBody:  errResp,
	}
}
