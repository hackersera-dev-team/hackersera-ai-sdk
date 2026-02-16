package hackeserasdk

// ─── Model Constants ────────────────────────────────────────────────────────

const (
	// ModelDefault is the default general-purpose model.
	ModelDefault = "hackersera-ai"
	// ModelPro is the advanced model for complex reasoning.
	ModelPro = "hackersera-ai-pro"
	// ModelLite is the fast, lightweight model.
	ModelLite = "hackersera-ai-lite"
	// ModelEmbedding is the model for text embeddings.
	ModelEmbedding = "hackersera-ai-embedding"
)

// ─── Chat Completions ───────────────────────────────────────────────────────

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model               string    `json:"model"`
	Messages            []Message `json:"messages"`
	Stream              bool      `json:"stream,omitempty"`
	MaxTokens           *int      `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int      `json:"max_completion_tokens,omitempty"`
	Temperature         *float64  `json:"temperature,omitempty"`
	TopP                *float64  `json:"top_p,omitempty"`
	Stop                []string  `json:"stop,omitempty"`
}

// IntPtr is a helper to create a pointer to an int value.
func IntPtr(v int) *int { return &v }

// Float64Ptr is a helper to create a pointer to a float64 value.
func Float64Ptr(v float64) *float64 { return &v }

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a non-streaming chat completion response.
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a single completion choice.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ─── Streaming ──────────────────────────────────────────────────────────────

// ChatStreamChunk represents a single SSE chunk during streaming.
type ChatStreamChunk struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChunkChoice `json:"choices"`
	Usage   *Usage        `json:"usage,omitempty"`
}

// ChunkChoice represents a single choice in a streaming chunk.
type ChunkChoice struct {
	Index        int     `json:"index"`
	Delta        Delta   `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

// Delta represents the incremental content in a streaming chunk.
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ─── Models ─────────────────────────────────────────────────────────────────

// Model represents a single model.
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelList represents the response from the models endpoint.
type ModelList struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// ─── Embeddings ─────────────────────────────────────────────────────────────

// EmbeddingRequest represents an embedding request.
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbeddingResponse represents the response from the embeddings endpoint.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

// EmbeddingData represents a single embedding vector.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingUsage represents token usage for embeddings.
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ─── Health ─────────────────────────────────────────────────────────────────

// HealthResponse represents the response from the health endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ─── Documents (RAG) ────────────────────────────────────────────────────────

// DocumentUploadRequest represents a single document upload request.
type DocumentUploadRequest struct {
	Content  string            `json:"content"`
	Filename string            `json:"filename,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
}

// DocumentBatchUploadRequest represents a batch document upload request.
type DocumentBatchUploadRequest struct {
	Documents []DocumentUploadRequest `json:"documents"`
}

// DocumentResponse represents a document returned by the API.
type DocumentResponse struct {
	ID         string            `json:"id"`
	Filename   string            `json:"filename"`
	Status     string            `json:"status"`
	ChunkCount int               `json:"chunk_count"`
	Tags       map[string]string `json:"tags,omitempty"`
	CreatedAt  string            `json:"created_at"`
	Error      string            `json:"error,omitempty"`
}

// DocumentListResponse represents the response from listing documents.
type DocumentListResponse struct {
	Object string             `json:"object"`
	Data   []DocumentResponse `json:"data"`
	Total  int                `json:"total"`
}

// DocumentDeleteResponse represents the response from deleting a document.
type DocumentDeleteResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// ─── Search (RAG) ───────────────────────────────────────────────────────────

// SearchRequest represents a semantic search request.
type SearchRequest struct {
	Query     string            `json:"query"`
	TopK      int               `json:"top_k,omitempty"`
	Threshold float64           `json:"threshold,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Filename   string  `json:"filename"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	ChunkIndex int     `json:"chunk_index"`
}

// SearchResponse represents the response from a search request.
type SearchResponse struct {
	Object string         `json:"object"`
	Data   []SearchResult `json:"data"`
	Query  string         `json:"query"`
	Total  int            `json:"total"`
}

// ─── Usage ──────────────────────────────────────────────────────────────────

// UsageByModel represents usage statistics for a single model.
type UsageByModel struct {
	Model            string  `json:"model"`
	Requests         int     `json:"requests"`
	TotalTokens      int     `json:"total_tokens"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	AvgLatencyMs     float64 `json:"avg_latency_ms"`
}

// UsageResponse represents the response from the usage endpoint.
type UsageResponse struct {
	TotalRequests    int            `json:"total_requests"`
	TotalTokens      int            `json:"total_tokens"`
	PromptTokens     int            `json:"prompt_tokens"`
	CompletionTokens int            `json:"completion_tokens"`
	AvgLatencyMs     float64        `json:"avg_latency_ms"`
	ByModel          []UsageByModel `json:"by_model"`
}

// UsageRecord represents a single usage record.
type UsageRecord struct {
	ID               int    `json:"id"`
	RequestID        string `json:"request_id"`
	Model            string `json:"model"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	LatencyMs        int64  `json:"latency_ms"`
	StatusCode       int    `json:"status_code"`
	Streaming        bool   `json:"streaming"`
	CreatedAt        string `json:"created_at"`
}

// UsageRecentResponse represents the response from the recent usage endpoint.
type UsageRecentResponse struct {
	Object string        `json:"object"`
	Count  int           `json:"count"`
	Data   []UsageRecord `json:"data"`
}

// ─── Cache Stats ────────────────────────────────────────────────────────────

// CacheStatsResponse represents the response from the cache stats endpoint.
type CacheStatsResponse struct {
	TotalEntries   int64   `json:"total_entries"`
	TotalHits      int64   `json:"total_hits"`
	ActiveEntries  int64   `json:"active_entries"`
	ExpiredEntries int64   `json:"expired_entries"`
	TokensSaved    int64   `json:"tokens_saved"`
	AvgHitCount    float64 `json:"avg_hit_count"`
	OldestEntry    string  `json:"oldest_entry,omitempty"`
	NewestEntry    string  `json:"newest_entry,omitempty"`
}

// ─── Readiness ──────────────────────────────────────────────────────────────

// ReadyResponse represents the response from the readiness endpoint.
type ReadyResponse struct {
	Ready   bool              `json:"ready"`
	Version string            `json:"version"`
	Checks  map[string]string `json:"checks"`
}

// ─── Errors ─────────────────────────────────────────────────────────────────

// APIError represents an error returned by the API.
type APIError struct {
	StatusCode int
	ErrorBody  ErrorResponse
}

func (e *APIError) Error() string {
	return e.ErrorBody.Error.Message
}

// ErrorResponse represents the JSON error body from the API.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the error message and type.
type ErrorDetail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param"`
	Code    *string `json:"code"`
}
