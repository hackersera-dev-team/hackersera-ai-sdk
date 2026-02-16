package hackeserasdk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client is the SDK client for the hackersera-ai-model-provider API.
type Client struct {
	baseURL           string
	apiKey            string
	httpClient        *http.Client
	userID            string
	conversationID    string
	cognitiveDisabled bool
}

// NewClient creates a new SDK client.
//
//	client := hackeserasdk.NewClient("https://api-ai.hackersera.com", "your-api-key")
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

// SetUserID sets the default X-User-ID header for all requests.
// Pass an empty string to clear.
func (c *Client) SetUserID(userID string) *Client {
	c.userID = userID
	return c
}

// SetConversationID sets the default X-Conversation-ID header for all requests.
// Pass an empty string to clear.
func (c *Client) SetConversationID(conversationID string) *Client {
	c.conversationID = conversationID
	return c
}

// SetCognitiveDisabled sets the default X-Cognitive-Disabled header for all requests.
func (c *Client) SetCognitiveDisabled(disabled bool) *Client {
	c.cognitiveDisabled = disabled
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

// ChatCompletionWithOptions sends a non-streaming chat completion request with per-request options.
// Options override the client-level defaults for this single request.
func (c *Client) ChatCompletionWithOptions(ctx context.Context, req ChatRequest, opts RequestOptions) (*ChatResponse, error) {
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
	applyOptions(httpReq, opts)

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

// ChatCompletionStreamWithOptions sends a streaming chat completion request with per-request options.
func (c *Client) ChatCompletionStreamWithOptions(ctx context.Context, req ChatRequest, opts RequestOptions) (<-chan ChatStreamChunk, <-chan error) {
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
		applyOptions(httpReq, opts)

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

			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

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

// ─── Conversations ──────────────────────────────────────────────────────────

// ListConversations returns a list of conversations.
// Use limit to control the number of results (default: 50).
func (c *Client) ListConversations(ctx context.Context, limit int) (*ConversationListResponse, error) {
	url := c.baseURL + "/v1/conversations"
	if limit > 0 {
		url += "?limit=" + strconv.Itoa(limit)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var listResp ConversationListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &listResp, nil
}

// GetConversation returns a conversation with all its turns.
func (c *Client) GetConversation(ctx context.Context, conversationID string) (*ConversationDetail, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/conversations/"+conversationID, nil)
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

	var detail ConversationDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &detail, nil
}

// SearchConversations performs a full-text search across all conversation turns.
func (c *Client) SearchConversations(ctx context.Context, query string, limit int) (*ConversationSearchResponse, error) {
	url := c.baseURL + "/v1/conversations/search?query=" + query
	if limit > 0 {
		url += "&limit=" + strconv.Itoa(limit)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var searchResp ConversationSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &searchResp, nil
}

// DeleteConversation deletes a conversation and all its turns.
func (c *Client) DeleteConversation(ctx context.Context, conversationID string) (*ConversationDeleteResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/v1/conversations/"+conversationID, nil)
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

	var delResp ConversationDeleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&delResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &delResp, nil
}

// ─── Feedback ───────────────────────────────────────────────────────────────

// SubmitFeedback submits feedback on an AI response.
// Positive feedback (rating: 1) reinforces good patterns; negative feedback (rating: -1)
// with corrections teaches the system what went wrong.
func (c *Client) SubmitFeedback(ctx context.Context, req FeedbackRequest) (*FeedbackResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/feedback", bytes.NewReader(body))
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

	var fbResp FeedbackResponse
	if err := json.NewDecoder(resp.Body).Decode(&fbResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &fbResp, nil
}

// ─── User Profiles ──────────────────────────────────────────────────────────

// GetProfile returns the user profile for the given user ID.
// Requires X-User-ID header — set via SetUserID() or pass opts.
func (c *Client) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/profile", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)
	httpReq.Header.Set("X-User-ID", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &profile, nil
}

// UpdateProfile updates the user profile for the given user ID.
// Preferences are merged — existing keys are updated, new keys are added.
func (c *Client) UpdateProfile(ctx context.Context, userID string, req ProfileUpdateRequest) (*UserProfile, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/v1/profile", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)
	httpReq.Header.Set("X-User-ID", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &profile, nil
}

// ─── Knowledge Graph ────────────────────────────────────────────────────────

// QueryKnowledgeGraph queries the knowledge graph for related concepts.
func (c *Client) QueryKnowledgeGraph(ctx context.Context, query string, limit int) (*KnowledgeGraphResponse, error) {
	url := c.baseURL + "/v1/knowledge/graph?query=" + query
	if limit > 0 {
		url += "&limit=" + strconv.Itoa(limit)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var graphResp KnowledgeGraphResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &graphResp, nil
}

// ─── Learned Facts ──────────────────────────────────────────────────────────

// ListFacts returns learned facts from the knowledge base.
// Set verified to non-nil to filter by verification status.
func (c *Client) ListFacts(ctx context.Context, limit int, verified *bool) (*FactListResponse, error) {
	url := c.baseURL + "/v1/knowledge/facts"
	sep := "?"
	if limit > 0 {
		url += sep + "limit=" + strconv.Itoa(limit)
		sep = "&"
	}
	if verified != nil {
		url += sep + "verified=" + strconv.FormatBool(*verified)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var factsResp FactListResponse
	if err := json.NewDecoder(resp.Body).Decode(&factsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &factsResp, nil
}

// CreateFact creates a single fact in the knowledge base.
func (c *Client) CreateFact(ctx context.Context, req FactCreateRequest) (*Fact, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/knowledge/facts", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var fact Fact
	if err := json.NewDecoder(resp.Body).Decode(&fact); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &fact, nil
}

// CreateFacts creates multiple facts in the knowledge base in a single request.
func (c *Client) CreateFacts(ctx context.Context, facts []FactCreateRequest) (*FactListResponse, error) {
	req := FactBatchCreateRequest{Facts: facts}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/knowledge/facts", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var factsResp FactListResponse
	if err := json.NewDecoder(resp.Body).Decode(&factsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &factsResp, nil
}

// UpdateFact updates an existing fact by ID.
// Only provided fields are updated.
func (c *Client) UpdateFact(ctx context.Context, factID int, req FactUpdateRequest) (*Fact, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/v1/knowledge/facts/"+strconv.Itoa(factID), bytes.NewReader(body))
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

	var fact Fact
	if err := json.NewDecoder(resp.Body).Decode(&fact); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &fact, nil
}

// ─── Cognitive Intelligence ─────────────────────────────────────────────────

// GetCognitiveStats returns system-wide cognitive statistics.
func (c *Client) GetCognitiveStats(ctx context.Context) (*CognitiveStatsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/cognitive/stats", nil)
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

	var stats CognitiveStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &stats, nil
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

// ─── Metrics ────────────────────────────────────────────────────────────────

// GetMetrics returns Prometheus metrics in text exposition format.
func (c *Client) GetMetrics(ctx context.Context) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/metrics", nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.parseError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	return string(body), nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if c.userID != "" {
		req.Header.Set("X-User-ID", c.userID)
	}
	if c.conversationID != "" {
		req.Header.Set("X-Conversation-ID", c.conversationID)
	}
	if c.cognitiveDisabled {
		req.Header.Set("X-Cognitive-Disabled", "true")
	}
}

func applyOptions(req *http.Request, opts RequestOptions) {
	if opts.UserID != "" {
		req.Header.Set("X-User-ID", opts.UserID)
	}
	if opts.ConversationID != "" {
		req.Header.Set("X-Conversation-ID", opts.ConversationID)
	}
	if opts.CognitiveDisabled {
		req.Header.Set("X-Cognitive-Disabled", "true")
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
