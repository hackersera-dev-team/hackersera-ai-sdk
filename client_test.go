package hackeserasdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ─── Helpers ────────────────────────────────────────────────────────────────

// newTestServer creates an httptest.Server that responds with the given status and body.
func newTestServer(t *testing.T, method, path string, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			t.Errorf("expected method %s, got %s", method, r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(r.URL.Path, path) && r.URL.Path != path {
			t.Errorf("expected path prefix %s, got %s", path, r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}))
}

// newTestServerFunc creates an httptest.Server with a custom handler.
func newTestServerFunc(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// ─── Chat Completions ───────────────────────────────────────────────────────

func TestChatCompletion(t *testing.T) {
	expected := ChatResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1700000000,
		Model:   "hackersera-ai",
		Choices: []Choice{
			{
				Index:        0,
				Message:      Message{Role: "assistant", Content: "Hello!"},
				FinishReason: "stop",
			},
		},
		Usage:          Usage{PromptTokens: 5, CompletionTokens: 2, TotalTokens: 7},
		ConversationID: "conv-abc123",
	}

	srv := newTestServer(t, http.MethodPost, "/v1/chat/completions", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	resp, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model:    ModelDefault,
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, resp.ID)
	}
	if resp.ConversationID != expected.ConversationID {
		t.Errorf("expected ConversationID %q, got %q", expected.ConversationID, resp.ConversationID)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello!" {
		t.Errorf("expected content %q, got %q", "Hello!", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 7 {
		t.Errorf("expected total tokens 7, got %d", resp.Usage.TotalTokens)
	}
}

func TestChatCompletionWithOptions(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers are set
		if r.Header.Get("X-User-ID") != "user-42" {
			t.Errorf("expected X-User-ID=user-42, got %q", r.Header.Get("X-User-ID"))
		}
		if r.Header.Get("X-Conversation-ID") != "conv-99" {
			t.Errorf("expected X-Conversation-ID=conv-99, got %q", r.Header.Get("X-Conversation-ID"))
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected auth header, got %q", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{
			ID:             "chatcmpl-opts",
			ConversationID: "conv-99",
			Choices:        []Choice{{Message: Message{Role: "assistant", Content: "OK"}}},
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	resp, err := client.ChatCompletionWithOptions(context.Background(), ChatRequest{
		Model:    ModelDefault,
		Messages: []Message{{Role: "user", Content: "test"}},
		User:     "user-42",
	}, RequestOptions{
		UserID:         "user-42",
		ConversationID: "conv-99",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ConversationID != "conv-99" {
		t.Errorf("expected ConversationID conv-99, got %q", resp.ConversationID)
	}
}

func TestChatCompletionWithCognitiveDisabled(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Cognitive-Disabled") != "true" {
			t.Errorf("expected X-Cognitive-Disabled=true, got %q", r.Header.Get("X-Cognitive-Disabled"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{
			ID:      "chatcmpl-cog",
			Choices: []Choice{{Message: Message{Role: "assistant", Content: "raw"}}},
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	_, err := client.ChatCompletionWithOptions(context.Background(), ChatRequest{
		Model:    ModelDefault,
		Messages: []Message{{Role: "user", Content: "test"}},
	}, RequestOptions{
		CognitiveDisabled: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatCompletionWithToolCalling(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		json.Unmarshal(body, &req)

		if len(req.Tools) != 1 {
			t.Errorf("expected 1 tool, got %d", len(req.Tools))
		}
		if req.Tools[0].Function.Name != "get_weather" {
			t.Errorf("expected tool name get_weather, got %q", req.Tools[0].Function.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{
			ID: "chatcmpl-tools",
			Choices: []Choice{{
				Message: Message{
					Role:    "assistant",
					Content: nil,
					ToolCalls: []ToolCall{{
						ID:   "call-1",
						Type: "function",
						Function: FunctionCall{
							Name:      "get_weather",
							Arguments: `{"location":"Tokyo"}`,
						},
					}},
				},
				FinishReason: "tool_calls",
			}},
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	resp, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model:    ModelDefault,
		Messages: []Message{{Role: "user", Content: "What is the weather in Tokyo?"}},
		Tools: []Tool{{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_weather",
				Description: "Get current weather",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{"type": "string"},
					},
					"required": []string{"location"},
				},
			},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Choices[0].FinishReason != "tool_calls" {
		t.Errorf("expected finish_reason tool_calls, got %q", resp.Choices[0].FinishReason)
	}
}

func TestChatCompletionWithAllParams(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatRequest
		json.Unmarshal(body, &req)

		if req.User != "user-1" {
			t.Errorf("expected user user-1, got %q", req.User)
		}
		if req.PresencePenalty == nil || *req.PresencePenalty != 0.5 {
			t.Errorf("expected presence_penalty 0.5")
		}
		if req.FrequencyPenalty == nil || *req.FrequencyPenalty != 0.3 {
			t.Errorf("expected frequency_penalty 0.3")
		}
		if req.Seed == nil || *req.Seed != 42 {
			t.Errorf("expected seed 42")
		}
		if req.ResponseFormat == nil || req.ResponseFormat.Type != "json_object" {
			t.Errorf("expected response_format json_object")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{
			ID:      "chatcmpl-params",
			Choices: []Choice{{Message: Message{Role: "assistant", Content: `{"answer":"ok"}`}}},
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	_, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model:            ModelDefault,
		Messages:         []Message{{Role: "user", Content: "test"}},
		User:             "user-1",
		PresencePenalty:  Float64Ptr(0.5),
		FrequencyPenalty: Float64Ptr(0.3),
		Seed:             IntPtr(42),
		ResponseFormat:   &ResponseFormat{Type: "json_object"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatCompletionStream(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`{"id":"chatcmpl-s1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"}}]}`,
			`{"id":"chatcmpl-s1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":" world"}}]}`,
			`{"id":"chatcmpl-s1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7}}`,
		}
		for _, c := range chunks {
			fmt.Fprintf(w, "data: %s\n\n", c)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	chunks, errs := client.ChatCompletionStream(context.Background(), ChatRequest{
		Model:    ModelDefault,
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})

	var content strings.Builder
	for chunk := range chunks {
		if len(chunk.Choices) > 0 {
			content.WriteString(chunk.Choices[0].Delta.Content)
		}
	}

	// Check for errors
	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected stream error: %v", err)
		}
	}

	if content.String() != "Hello world" {
		t.Errorf("expected streamed content %q, got %q", "Hello world", content.String())
	}
}

func TestChatCompletionStreamWithOptions(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-ID") != "stream-user" {
			t.Errorf("expected X-User-ID=stream-user, got %q", r.Header.Get("X-User-ID"))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "data: %s\n\n", `{"id":"s1","choices":[{"index":0,"delta":{"content":"ok"}}]}`)
		fmt.Fprint(w, "data: [DONE]\n\n")
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	chunks, errs := client.ChatCompletionStreamWithOptions(context.Background(), ChatRequest{
		Model:    ModelDefault,
		Messages: []Message{{Role: "user", Content: "test"}},
	}, RequestOptions{UserID: "stream-user"})

	for range chunks {
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

// ─── Error Handling ─────────────────────────────────────────────────────────

func TestAPIError(t *testing.T) {
	errBody := ErrorResponse{
		Error: ErrorDetail{
			Message: "Missing Authorization header",
			Type:    "invalid_request_error",
			Code:    StringPtr("invalid_api_key"),
		},
	}

	srv := newTestServer(t, http.MethodGet, "/v1/models", http.StatusUnauthorized, errBody)
	defer srv.Close()

	client := NewClient(srv.URL, "")
	_, err := client.ListModels(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	if apiErr.Error() != "Missing Authorization header" {
		t.Errorf("expected error message %q, got %q", "Missing Authorization header", apiErr.Error())
	}
	if apiErr.ErrorBody.Error.Type != "invalid_request_error" {
		t.Errorf("expected error type %q, got %q", "invalid_request_error", apiErr.ErrorBody.Error.Type)
	}
}

func TestAPIErrorNonJSON(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	_, err := client.ListModels(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
	if apiErr.ErrorBody.Error.Type != "unknown_error" {
		t.Errorf("expected error type unknown_error, got %q", apiErr.ErrorBody.Error.Type)
	}
}

// ─── Client Configuration ───────────────────────────────────────────────────

func TestSetHeaders(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer my-key" {
			t.Errorf("expected auth header, got %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected content-type, got %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-User-ID") != "global-user" {
			t.Errorf("expected X-User-ID=global-user, got %q", r.Header.Get("X-User-ID"))
		}
		if r.Header.Get("X-Conversation-ID") != "global-conv" {
			t.Errorf("expected X-Conversation-ID=global-conv, got %q", r.Header.Get("X-Conversation-ID"))
		}
		if r.Header.Get("X-Cognitive-Disabled") != "true" {
			t.Errorf("expected X-Cognitive-Disabled=true, got %q", r.Header.Get("X-Cognitive-Disabled"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ModelList{Object: "list", Data: []Model{{ID: "hackersera-ai"}}})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "my-key").
		SetUserID("global-user").
		SetConversationID("global-conv").
		SetCognitiveDisabled(true)

	// Use ListModels (not Health) because Health doesn't call setHeaders by design
	_, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	client := NewClient("https://api.example.com/", "key")
	if client.baseURL != "https://api.example.com" {
		t.Errorf("expected trimmed URL, got %q", client.baseURL)
	}
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	client := NewClient("http://localhost", "key").WithHTTPClient(custom)
	if client.httpClient != custom {
		t.Error("expected custom HTTP client to be set")
	}
}

// ─── Models ─────────────────────────────────────────────────────────────────

func TestListModels(t *testing.T) {
	expected := ModelList{
		Object: "list",
		Data: []Model{
			{ID: "hackersera-ai", Object: "model", OwnedBy: "hackersera"},
			{ID: "hackersera-ai-pro", Object: "model", OwnedBy: "hackersera"},
		},
	}

	srv := newTestServer(t, http.MethodGet, "/v1/models", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	models, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models.Data) != 2 {
		t.Errorf("expected 2 models, got %d", len(models.Data))
	}
	if models.Data[0].ID != "hackersera-ai" {
		t.Errorf("expected model ID hackersera-ai, got %q", models.Data[0].ID)
	}
}

func TestGetModel(t *testing.T) {
	expected := Model{ID: "hackersera-ai", Object: "model", OwnedBy: "hackersera"}

	srv := newTestServer(t, http.MethodGet, "/v1/models/", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	model, err := client.GetModel(context.Background(), "hackersera-ai")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ID != "hackersera-ai" {
		t.Errorf("expected model ID hackersera-ai, got %q", model.ID)
	}
}

// ─── Embeddings ─────────────────────────────────────────────────────────────

func TestCreateEmbedding(t *testing.T) {
	expected := EmbeddingResponse{
		Object: "list",
		Data: []EmbeddingData{
			{Object: "embedding", Embedding: []float64{0.1, 0.2, 0.3}, Index: 0},
		},
		Model: "text-embedding-ada-002",
		Usage: EmbeddingUsage{PromptTokens: 2, TotalTokens: 2},
	}

	srv := newTestServer(t, http.MethodPost, "/v1/embeddings", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	resp, err := client.CreateEmbedding(context.Background(), EmbeddingRequest{
		Input: "Hello world",
		Model: ModelEmbedding,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 embedding, got %d", len(resp.Data))
	}
	if len(resp.Data[0].Embedding) != 3 {
		t.Errorf("expected 3 dimensions, got %d", len(resp.Data[0].Embedding))
	}
}

func TestCreateEmbeddingWithDimensions(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var raw map[string]interface{}
		json.Unmarshal(body, &raw)

		if raw["dimensions"] == nil {
			t.Error("expected dimensions field in request")
		}
		if int(raw["dimensions"].(float64)) != 768 {
			t.Errorf("expected dimensions 768, got %v", raw["dimensions"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(EmbeddingResponse{
			Object: "list",
			Data:   []EmbeddingData{{Embedding: make([]float64, 768)}},
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	_, err := client.CreateEmbedding(context.Background(), EmbeddingRequest{
		Input:      []string{"Hello", "World"},
		Model:      ModelEmbedding,
		Dimensions: IntPtr(768),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── Health ─────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	expected := HealthResponse{Status: "ok", Version: "1.1.5"}

	srv := newTestServer(t, http.MethodGet, "/health", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != "ok" {
		t.Errorf("expected status ok, got %q", health.Status)
	}
	if health.Version != "1.1.5" {
		t.Errorf("expected version 1.1.5, got %q", health.Version)
	}
}

func TestHealthDegraded(t *testing.T) {
	expected := HealthResponse{Status: "degraded", Version: "1.1.5"}

	srv := newTestServer(t, http.MethodGet, "/health", http.StatusServiceUnavailable, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != "degraded" {
		t.Errorf("expected status degraded, got %q", health.Status)
	}
}

// ─── Ready ──────────────────────────────────────────────────────────────────

func TestReady(t *testing.T) {
	expected := ReadyResponse{
		Ready:   true,
		Version: "1.1.5",
		Checks:  map[string]string{"backend": "ok", "database": "ok"},
	}

	srv := newTestServer(t, http.MethodGet, "/ready", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	ready, err := client.Ready(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ready.Ready {
		t.Error("expected ready=true")
	}
	if ready.Checks["backend"] != "ok" {
		t.Errorf("expected backend=ok, got %q", ready.Checks["backend"])
	}
}

// ─── Documents (RAG) ────────────────────────────────────────────────────────

func TestUploadDocument(t *testing.T) {
	expected := DocumentResponse{
		ID:       "doc-abc123",
		Filename: "test.md",
		Status:   "processing",
		Tags:     map[string]string{"topic": "test"},
	}

	srv := newTestServer(t, http.MethodPost, "/v1/documents", http.StatusAccepted, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	doc, err := client.UploadDocument(context.Background(), DocumentUploadRequest{
		Content:  "Test content",
		Filename: "test.md",
		Tags:     map[string]string{"topic": "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.ID != "doc-abc123" {
		t.Errorf("expected doc ID doc-abc123, got %q", doc.ID)
	}
	if doc.Status != "processing" {
		t.Errorf("expected status processing, got %q", doc.Status)
	}
}

func TestUploadDocuments(t *testing.T) {
	expected := DocumentListResponse{
		Object: "list",
		Data: []DocumentResponse{
			{ID: "doc-1", Filename: "a.md", Status: "processing"},
			{ID: "doc-2", Filename: "b.md", Status: "processing"},
		},
		Total: 2,
	}

	srv := newTestServer(t, http.MethodPost, "/v1/documents", http.StatusAccepted, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	resp, err := client.UploadDocuments(context.Background(), []DocumentUploadRequest{
		{Content: "Doc 1", Filename: "a.md"},
		{Content: "Doc 2", Filename: "b.md"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
}

func TestListDocuments(t *testing.T) {
	expected := DocumentListResponse{
		Object: "list",
		Data:   []DocumentResponse{{ID: "doc-1", Filename: "test.md", Status: "indexed"}},
		Total:  1,
	}

	srv := newTestServer(t, http.MethodGet, "/v1/documents", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	docs, err := client.ListDocuments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docs.Total != 1 {
		t.Errorf("expected total 1, got %d", docs.Total)
	}
}

func TestGetDocument(t *testing.T) {
	expected := DocumentResponse{
		ID:         "doc-abc",
		Filename:   "test.md",
		Status:     "indexed",
		ChunkCount: 5,
	}

	srv := newTestServer(t, http.MethodGet, "/v1/documents/", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	doc, err := client.GetDocument(context.Background(), "doc-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.ChunkCount != 5 {
		t.Errorf("expected chunk count 5, got %d", doc.ChunkCount)
	}
}

func TestDeleteDocument(t *testing.T) {
	expected := DocumentDeleteResponse{ID: "doc-abc", Deleted: true}

	srv := newTestServer(t, http.MethodDelete, "/v1/documents/", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	del, err := client.DeleteDocument(context.Background(), "doc-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !del.Deleted {
		t.Error("expected deleted=true")
	}
}

// ─── Search (RAG) ───────────────────────────────────────────────────────────

func TestSearch(t *testing.T) {
	expected := SearchResponse{
		Object: "list",
		Data: []SearchResult{
			{ChunkID: "chunk-1", DocumentID: "doc-1", Filename: "test.md", Content: "result", Score: 0.87, ChunkIndex: 0},
		},
		Query: "test query",
		Total: 1,
	}

	srv := newTestServer(t, http.MethodPost, "/v1/search", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	resp, err := client.Search(context.Background(), SearchRequest{
		Query:     "test query",
		TopK:      5,
		Threshold: 0.3,
		Tags:      map[string]string{"topic": "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
	if resp.Data[0].Score != 0.87 {
		t.Errorf("expected score 0.87, got %f", resp.Data[0].Score)
	}
}

// ─── Conversations ──────────────────────────────────────────────────────────

func TestListConversations(t *testing.T) {
	expected := ConversationListResponse{
		Object: "list",
		Data: []Conversation{
			{ID: "conv-1", Title: "Docker question", TurnCount: 4, Model: "glm-4.7"},
			{ID: "conv-2", Title: "Go channels", TurnCount: 2, Model: "glm-4.7"},
		},
		Total: 2,
	}

	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	convos, err := client.ListConversations(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if convos.Total != 2 {
		t.Errorf("expected total 2, got %d", convos.Total)
	}
	if convos.Data[0].Title != "Docker question" {
		t.Errorf("expected title %q, got %q", "Docker question", convos.Data[0].Title)
	}
}

func TestListConversationsNoLimit(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ConversationListResponse{Object: "list", Total: 0})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	_, err := client.ListConversations(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetConversation(t *testing.T) {
	expected := ConversationDetail{
		ID:        "conv-1",
		Title:     "Docker question",
		TurnCount: 2,
		Turns: []ConversationTurn{
			{ID: 1, Role: "user", Content: "What is Docker?"},
			{ID: 2, Role: "assistant", Content: "Docker is a platform..."},
		},
	}

	srv := newTestServer(t, http.MethodGet, "/v1/conversations/", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	detail, err := client.GetConversation(context.Background(), "conv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detail.Turns) != 2 {
		t.Errorf("expected 2 turns, got %d", len(detail.Turns))
	}
	if detail.Turns[0].Content != "What is Docker?" {
		t.Errorf("expected first turn content, got %q", detail.Turns[0].Content)
	}
}

func TestSearchConversations(t *testing.T) {
	expected := ConversationSearchResponse{
		Object: "list",
		Data: []ConversationSearchResult{
			{ConversationID: "conv-1", TurnID: 1, Role: "user", Content: "What is Docker?"},
		},
		Query: "docker",
		Total: 1,
	}

	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "docker" {
			t.Errorf("expected query=docker, got %q", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("limit") != "20" {
			t.Errorf("expected limit=20, got %q", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	resp, err := client.SearchConversations(context.Background(), "docker", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
}

func TestDeleteConversation(t *testing.T) {
	expected := ConversationDeleteResponse{ID: "conv-1", Deleted: true}

	srv := newTestServer(t, http.MethodDelete, "/v1/conversations/", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	del, err := client.DeleteConversation(context.Background(), "conv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !del.Deleted {
		t.Error("expected deleted=true")
	}
	if del.ID != "conv-1" {
		t.Errorf("expected ID conv-1, got %q", del.ID)
	}
}

// ─── Feedback ───────────────────────────────────────────────────────────────

func TestSubmitFeedback(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req FeedbackRequest
		json.Unmarshal(body, &req)

		if req.ConversationID != "conv-1" {
			t.Errorf("expected conversation_id conv-1, got %q", req.ConversationID)
		}
		if req.Rating != 1 {
			t.Errorf("expected rating 1, got %d", req.Rating)
		}
		if req.Comment != "Great answer" {
			t.Errorf("expected comment %q, got %q", "Great answer", req.Comment)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FeedbackResponse{
			ID:             7,
			ConversationID: "conv-1",
			TurnID:         6,
			Rating:         1,
			CreatedAt:      "2026-02-16T12:08:15Z",
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	fb, err := client.SubmitFeedback(context.Background(), FeedbackRequest{
		ConversationID: "conv-1",
		TurnID:         6,
		Rating:         1,
		Comment:        "Great answer",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.ID != 7 {
		t.Errorf("expected feedback ID 7, got %d", fb.ID)
	}
	if fb.Rating != 1 {
		t.Errorf("expected rating 1, got %d", fb.Rating)
	}
}

func TestSubmitNegativeFeedbackWithCorrection(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req FeedbackRequest
		json.Unmarshal(body, &req)

		if req.Rating != -1 {
			t.Errorf("expected rating -1, got %d", req.Rating)
		}
		if req.Correction != "The correct answer is..." {
			t.Errorf("expected correction, got %q", req.Correction)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FeedbackResponse{ID: 8, Rating: -1})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	fb, err := client.SubmitFeedback(context.Background(), FeedbackRequest{
		ConversationID: "conv-1",
		Rating:         -1,
		Correction:     "The correct answer is...",
		ChunkIDs:       []string{"chunk-1", "chunk-2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.Rating != -1 {
		t.Errorf("expected rating -1, got %d", fb.Rating)
	}
}

// ─── User Profiles ──────────────────────────────────────────────────────────

func TestGetProfile(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-User-ID") != "user-123" {
			t.Errorf("expected X-User-ID=user-123, got %q", r.Header.Get("X-User-ID"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserProfile{
			UserID:       "user-123",
			DisplayName:  "John",
			Preferences:  map[string]string{"language": "go"},
			Expertise:    map[string]float64{"docker": 0.85, "go": 0.45},
			Topics:       map[string]int{"containers": 12},
			TotalQueries: 23,
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	profile, err := client.GetProfile(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UserID != "user-123" {
		t.Errorf("expected user_id user-123, got %q", profile.UserID)
	}
	if profile.TotalQueries != 23 {
		t.Errorf("expected 23 queries, got %d", profile.TotalQueries)
	}
	if profile.Expertise["docker"] != 0.85 {
		t.Errorf("expected docker expertise 0.85, got %f", profile.Expertise["docker"])
	}
}

func TestUpdateProfile(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get("X-User-ID") != "user-123" {
			t.Errorf("expected X-User-ID=user-123, got %q", r.Header.Get("X-User-ID"))
		}

		body, _ := io.ReadAll(r.Body)
		var req ProfileUpdateRequest
		json.Unmarshal(body, &req)

		if req.DisplayName != "John Doe" {
			t.Errorf("expected display_name John Doe, got %q", req.DisplayName)
		}
		if req.Preferences["language"] != "go" {
			t.Errorf("expected language=go, got %q", req.Preferences["language"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UserProfile{
			UserID:      "user-123",
			DisplayName: "John Doe",
			Preferences: map[string]string{"language": "go", "detail_level": "detailed"},
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	profile, err := client.UpdateProfile(context.Background(), "user-123", ProfileUpdateRequest{
		DisplayName: "John Doe",
		Preferences: map[string]string{"language": "go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.DisplayName != "John Doe" {
		t.Errorf("expected display_name John Doe, got %q", profile.DisplayName)
	}
}

// ─── Knowledge Graph ────────────────────────────────────────────────────────

func TestQueryKnowledgeGraph(t *testing.T) {
	expected := KnowledgeGraphResponse{
		Object: "list",
		Data: []KnowledgeNode{
			{ID: "node-1", Label: "containers", Type: "concept", HitCount: 12},
			{ID: "node-2", Label: "kubernetes", Type: "concept", HitCount: 8},
		},
		Edges: []KnowledgeEdge{
			{ID: 46, FromID: "node-2", ToID: "node-1", Relation: "co_queried", Weight: 1.0},
		},
		Query: "docker",
		Total: 2,
	}

	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "docker" {
			t.Errorf("expected query=docker, got %q", r.URL.Query().Get("query"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	graph, err := client.QueryKnowledgeGraph(context.Background(), "docker", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if graph.Total != 2 {
		t.Errorf("expected total 2, got %d", graph.Total)
	}
	if len(graph.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(graph.Edges))
	}
	if graph.Edges[0].Relation != "co_queried" {
		t.Errorf("expected relation co_queried, got %q", graph.Edges[0].Relation)
	}
}

// ─── Learned Facts ──────────────────────────────────────────────────────────

func TestListFacts(t *testing.T) {
	expected := FactListResponse{
		Object: "list",
		Data: []Fact{
			{ID: 1, Content: "Docker uses cgroups", Source: "conversation", Confidence: 0.8, Verified: false},
			{ID: 2, Content: "Go 1.23 supports range over integers", Source: "manual", Confidence: 0.95, Verified: true},
		},
		Total: 2,
	}

	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "20" {
			t.Errorf("expected limit=20, got %q", r.URL.Query().Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	facts, err := client.ListFacts(context.Background(), 20, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if facts.Total != 2 {
		t.Errorf("expected total 2, got %d", facts.Total)
	}
}

func TestListFactsVerifiedFilter(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("verified") != "true" {
			t.Errorf("expected verified=true, got %q", r.URL.Query().Get("verified"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FactListResponse{Object: "list", Total: 1})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	_, err := client.ListFacts(context.Background(), 10, BoolPtr(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateFact(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req FactCreateRequest
		json.Unmarshal(body, &req)

		if req.Content != "Go is awesome" {
			t.Errorf("expected content %q, got %q", "Go is awesome", req.Content)
		}
		if req.Source != "manual" {
			t.Errorf("expected source manual, got %q", req.Source)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Fact{
			ID:         14,
			Content:    req.Content,
			Source:     req.Source,
			Confidence: req.Confidence,
			Verified:   req.Verified,
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	fact, err := client.CreateFact(context.Background(), FactCreateRequest{
		Content:    "Go is awesome",
		Source:     "manual",
		Confidence: 0.9,
		Verified:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fact.ID != 14 {
		t.Errorf("expected fact ID 14, got %d", fact.ID)
	}
}

func TestCreateFacts(t *testing.T) {
	expected := FactListResponse{
		Object: "list",
		Data: []Fact{
			{ID: 15, Content: "Fact 1"},
			{ID: 16, Content: "Fact 2"},
		},
		Total: 2,
	}

	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req FactBatchCreateRequest
		json.Unmarshal(body, &req)

		if len(req.Facts) != 2 {
			t.Errorf("expected 2 facts, got %d", len(req.Facts))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	resp, err := client.CreateFacts(context.Background(), []FactCreateRequest{
		{Content: "Fact 1", Source: "docs"},
		{Content: "Fact 2", Source: "docs"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
}

func TestUpdateFact(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/15") {
			t.Errorf("expected path ending in /15, got %q", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var raw map[string]interface{}
		json.Unmarshal(body, &raw)

		if raw["verified"] != true {
			t.Errorf("expected verified=true")
		}
		if raw["confidence"].(float64) != 0.99 {
			t.Errorf("expected confidence=0.99")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Fact{
			ID:         15,
			Content:    "Updated content",
			Confidence: 0.99,
			Verified:   true,
		})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	fact, err := client.UpdateFact(context.Background(), 15, FactUpdateRequest{
		Verified:   BoolPtr(true),
		Confidence: Float64Ptr(0.99),
		Content:    StringPtr("Updated content"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fact.Confidence != 0.99 {
		t.Errorf("expected confidence 0.99, got %f", fact.Confidence)
	}
	if !fact.Verified {
		t.Error("expected verified=true")
	}
}

// ─── Cognitive Intelligence ─────────────────────────────────────────────────

func TestGetCognitiveStats(t *testing.T) {
	expected := CognitiveStatsResponse{
		TotalConversations:  114,
		TotalTurns:          228,
		TotalFeedback:       8,
		PositiveFeedback:    4,
		NegativeFeedback:    4,
		TotalUsers:          1,
		TotalKnowledgeNodes: 75,
		TotalKnowledgeEdges: 437,
		TotalLearnedFacts:   17,
		VerifiedFacts:       4,
		AvgFactConfidence:   0.755,
	}

	srv := newTestServer(t, http.MethodGet, "/v1/cognitive/stats", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	stats, err := client.GetCognitiveStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalConversations != 114 {
		t.Errorf("expected 114 conversations, got %d", stats.TotalConversations)
	}
	if stats.TotalKnowledgeNodes != 75 {
		t.Errorf("expected 75 nodes, got %d", stats.TotalKnowledgeNodes)
	}
	if stats.AvgFactConfidence != 0.755 {
		t.Errorf("expected avg confidence 0.755, got %f", stats.AvgFactConfidence)
	}
}

// ─── Usage ──────────────────────────────────────────────────────────────────

func TestGetUsage(t *testing.T) {
	expected := UsageResponse{
		TotalRequests:    100,
		TotalTokens:      50000,
		PromptTokens:     30000,
		CompletionTokens: 20000,
		AvgLatencyMs:     1500.5,
		ByModel: []UsageByModel{
			{Model: "hackersera-ai", Requests: 80, TotalTokens: 40000},
		},
	}

	srv := newTestServer(t, http.MethodGet, "/v1/usage", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	usage, err := client.GetUsage(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.TotalRequests != 100 {
		t.Errorf("expected 100 requests, got %d", usage.TotalRequests)
	}
}

func TestGetRecentUsage(t *testing.T) {
	expected := UsageRecentResponse{
		Object: "list",
		Count:  2,
		Data: []UsageRecord{
			{ID: 1, RequestID: "req-1", Model: "hackersera-ai", TotalTokens: 50},
			{ID: 2, RequestID: "req-2", Model: "hackersera-ai", TotalTokens: 30},
		},
	}

	srv := newTestServer(t, http.MethodGet, "/v1/usage/recent", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	recent, err := client.GetRecentUsage(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recent.Count != 2 {
		t.Errorf("expected count 2, got %d", recent.Count)
	}
}

// ─── Cache Stats ────────────────────────────────────────────────────────────

func TestGetCacheStats(t *testing.T) {
	expected := CacheStatsResponse{
		TotalEntries:  100,
		TotalHits:     50,
		ActiveEntries: 80,
		TokensSaved:   10000,
		AvgHitCount:   2.5,
	}

	srv := newTestServer(t, http.MethodGet, "/v1/cache/stats", http.StatusOK, expected)
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	stats, err := client.GetCacheStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalHits != 50 {
		t.Errorf("expected 50 hits, got %d", stats.TotalHits)
	}
	if stats.TokensSaved != 10000 {
		t.Errorf("expected 10000 tokens saved, got %d", stats.TokensSaved)
	}
}

// ─── Metrics ────────────────────────────────────────────────────────────────

func TestGetMetrics(t *testing.T) {
	metricsBody := `# HELP hackersera_uptime_seconds Time since server start
# TYPE hackersera_uptime_seconds gauge
hackersera_uptime_seconds 3600
# HELP hackersera_http_requests_total Total HTTP requests
# TYPE hackersera_http_requests_total counter
hackersera_http_requests_total{method="POST",path="/v1/chat/completions",status="200"} 42
`

	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/metrics" {
			t.Errorf("expected path /metrics, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(metricsBody))
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	metrics, err := client.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(metrics, "hackersera_uptime_seconds") {
		t.Error("expected metrics to contain hackersera_uptime_seconds")
	}
	if !strings.Contains(metrics, "hackersera_http_requests_total") {
		t.Error("expected metrics to contain hackersera_http_requests_total")
	}
}

// ─── Helper Functions ───────────────────────────────────────────────────────

func TestHelperFunctions(t *testing.T) {
	i := IntPtr(42)
	if *i != 42 {
		t.Errorf("IntPtr: expected 42, got %d", *i)
	}

	f := Float64Ptr(3.14)
	if *f != 3.14 {
		t.Errorf("Float64Ptr: expected 3.14, got %f", *f)
	}

	b := BoolPtr(true)
	if !*b {
		t.Error("BoolPtr: expected true")
	}

	s := StringPtr("hello")
	if *s != "hello" {
		t.Errorf("StringPtr: expected hello, got %q", *s)
	}
}

// ─── Context Cancellation ───────────────────────────────────────────────────

func TestContextCancellation(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response — never responds
		select {}
	})
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.ListModels(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// ─── Request Body Validation ────────────────────────────────────────────────

func TestChatRequestSerialization(t *testing.T) {
	req := ChatRequest{
		Model: ModelDefault,
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "Hello"},
		},
		Temperature:      Float64Ptr(0.7),
		MaxTokens:        IntPtr(100),
		TopP:             Float64Ptr(0.9),
		Stop:             []string{"\n"},
		PresencePenalty:  Float64Ptr(0.5),
		FrequencyPenalty: Float64Ptr(0.3),
		User:             "user-1",
		Seed:             IntPtr(42),
		ResponseFormat:   &ResponseFormat{Type: "json_object"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if raw["model"] != "hackersera-ai" {
		t.Errorf("expected model hackersera-ai, got %v", raw["model"])
	}
	if raw["temperature"].(float64) != 0.7 {
		t.Errorf("expected temperature 0.7, got %v", raw["temperature"])
	}
	if raw["presence_penalty"].(float64) != 0.5 {
		t.Errorf("expected presence_penalty 0.5, got %v", raw["presence_penalty"])
	}
	if raw["user"] != "user-1" {
		t.Errorf("expected user user-1, got %v", raw["user"])
	}
	if raw["seed"].(float64) != 42 {
		t.Errorf("expected seed 42, got %v", raw["seed"])
	}

	rf := raw["response_format"].(map[string]interface{})
	if rf["type"] != "json_object" {
		t.Errorf("expected response_format type json_object, got %v", rf["type"])
	}
}

func TestFeedbackRequestSerialization(t *testing.T) {
	req := FeedbackRequest{
		ConversationID: "conv-1",
		TurnID:         6,
		Rating:         -1,
		Comment:        "Wrong answer",
		Correction:     "The correct answer is X",
		ChunkIDs:       []string{"chunk-a", "chunk-b"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if raw["conversation_id"] != "conv-1" {
		t.Errorf("expected conversation_id conv-1, got %v", raw["conversation_id"])
	}
	if raw["rating"].(float64) != -1 {
		t.Errorf("expected rating -1, got %v", raw["rating"])
	}
	if raw["correction"] != "The correct answer is X" {
		t.Errorf("expected correction, got %v", raw["correction"])
	}

	chunkIDs := raw["chunk_ids"].([]interface{})
	if len(chunkIDs) != 2 {
		t.Errorf("expected 2 chunk_ids, got %d", len(chunkIDs))
	}
}

func TestFactUpdateRequestOmitsNil(t *testing.T) {
	// Only set verified, leave others nil
	req := FactUpdateRequest{
		Verified: BoolPtr(true),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, exists := raw["content"]; exists {
		t.Error("expected content to be omitted when nil")
	}
	if _, exists := raw["confidence"]; exists {
		t.Error("expected confidence to be omitted when nil")
	}
	if raw["verified"] != true {
		t.Errorf("expected verified=true, got %v", raw["verified"])
	}
}

// ─── No Auth Header When Key Empty ──────────────────────────────────────────

func TestNoAuthHeaderWhenKeyEmpty(t *testing.T) {
	srv := newTestServerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("expected no Authorization header, got %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	})
	defer srv.Close()

	client := NewClient(srv.URL, "")
	_, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── Model Constants ────────────────────────────────────────────────────────

func TestModelConstants(t *testing.T) {
	if ModelDefault != "hackersera-ai" {
		t.Errorf("expected ModelDefault=hackersera-ai, got %q", ModelDefault)
	}
	if ModelPro != "hackersera-ai-pro" {
		t.Errorf("expected ModelPro=hackersera-ai-pro, got %q", ModelPro)
	}
	if ModelLite != "hackersera-ai-lite" {
		t.Errorf("expected ModelLite=hackersera-ai-lite, got %q", ModelLite)
	}
	if ModelEmbedding != "hackersera-ai-embedding" {
		t.Errorf("expected ModelEmbedding=hackersera-ai-embedding, got %q", ModelEmbedding)
	}
}
