package hackeserasdk

// ─── Chat Completions ───────────────────────────────────────────────────────

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a non-streaming chat completion response.
type ChatResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
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
	Status    string `json:"status"`
	ClaudeCLI string `json:"claude_cli"`
	Version   string `json:"version"`
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
