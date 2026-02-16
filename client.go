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

// ─── Documents (RAG) ────────────────────────────────────────────────────────

// UploadDocument uploads a single document for RAG ingestion.
// Returns immediately with status "processing" (202 Accepted); ingestion is async.
// Poll with GetDocument() to check when indexing completes.
func (c *Client) UploadDocument(ctx context.Context, req DocumentUploadRequest) (*DocumentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/documents", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var docResp DocumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&docResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &docResp, nil
}

// UploadDocuments uploads multiple documents for RAG ingestion in a single request.
// Returns immediately with status "processing" (202 Accepted); ingestion is async.
func (c *Client) UploadDocuments(ctx context.Context, docs []DocumentUploadRequest) (*DocumentListResponse, error) {
	req := DocumentBatchUploadRequest{Documents: docs}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/documents", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var listResp DocumentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &listResp, nil
}

// ListDocuments returns all documents in the knowledge base.
func (c *Client) ListDocuments(ctx context.Context) (*DocumentListResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/documents", nil)
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

	var listResp DocumentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &listResp, nil
}

// GetDocument returns a single document by ID.
// Use this to poll document status after uploading.
func (c *Client) GetDocument(ctx context.Context, docID string) (*DocumentResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/documents/"+docID, nil)
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

	var docResp DocumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&docResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &docResp, nil
}

// DeleteDocument soft-deletes a document and removes its chunks.
func (c *Client) DeleteDocument(ctx context.Context, docID string) (*DocumentDeleteResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/v1/documents/"+docID, nil)
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

	var delResp DocumentDeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&delResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &delResp, nil
}

// ─── Search (RAG) ───────────────────────────────────────────────────────────

// Search performs a semantic search over the knowledge base.
// Uses hybrid search (pgvector cosine + keyword RRF) for best results.
func (c *Client) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/search", bytes.NewReader(body))
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

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &searchResp, nil
}

// ─── Usage ──────────────────────────────────────────────────────────────────

// GetUsage returns aggregated usage statistics.
func (c *Client) GetUsage(ctx context.Context) (*UsageResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/usage", nil)
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

	var usageResp UsageResponse
	if err := json.NewDecoder(resp.Body).Decode(&usageResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &usageResp, nil
}

// GetRecentUsage returns recent usage records.
func (c *Client) GetRecentUsage(ctx context.Context) (*UsageRecentResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/usage/recent", nil)
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

	var recentResp UsageRecentResponse
	if err := json.NewDecoder(resp.Body).Decode(&recentResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &recentResp, nil
}

// ─── Cache Stats ────────────────────────────────────────────────────────────

// GetCacheStats returns response cache statistics.
func (c *Client) GetCacheStats(ctx context.Context) (*CacheStatsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/cache/stats", nil)
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

	var statsResp CacheStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&statsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &statsResp, nil
}

// ─── Readiness ──────────────────────────────────────────────────────────────

// Ready checks if the server is ready to accept requests (database + backend connected).
func (c *Client) Ready(ctx context.Context) (*ReadyResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/ready", nil)
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

	var readyResp ReadyResponse
	if err := json.NewDecoder(resp.Body).Decode(&readyResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &readyResp, nil
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
