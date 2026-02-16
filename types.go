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

// ─── Helpers ────────────────────────────────────────────────────────────────

// IntPtr is a helper to create a pointer to an int value.
func IntPtr(v int) *int { return &v }

// Float64Ptr is a helper to create a pointer to a float64 value.
func Float64Ptr(v float64) *float64 { return &v }

// BoolPtr is a helper to create a pointer to a bool value.
func BoolPtr(v bool) *bool { return &v }

// StringPtr is a helper to create a pointer to a string value.
func StringPtr(v string) *string { return &v }

// ─── Chat Completions ───────────────────────────────────────────────────────

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model               string          `json:"model"`
	Messages            []Message       `json:"messages"`
	Stream              bool            `json:"stream,omitempty"`
	MaxTokens           *int            `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"`
	Temperature         *float64        `json:"temperature,omitempty"`
	TopP                *float64        `json:"top_p,omitempty"`
	Stop                []string        `json:"stop,omitempty"`
	PresencePenalty     *float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty    *float64        `json:"frequency_penalty,omitempty"`
	User                string          `json:"user,omitempty"`
	Tools               []Tool          `json:"tools,omitempty"`
	ToolChoice          interface{}     `json:"tool_choice,omitempty"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
	Seed                *int            `json:"seed,omitempty"`
}

// Message represents a single message in a conversation.
type Message struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"`
	Name       string      `json:"name,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

// ContentPart represents a single part of a multimodal content array.
type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL in a multimodal content part.
type ImageURL struct {
	URL string `json:"url"`
}

// Tool represents a tool definition for function calling.
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents the function definition within a tool.
type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// ToolCall represents a tool call made by the assistant.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents the function name and arguments in a tool call.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ResponseFormat specifies the desired response format.
type ResponseFormat struct {
	Type string `json:"type"`
}

// ChatResponse represents a non-streaming chat completion response.
type ChatResponse struct {
	ID             string   `json:"id"`
	Object         string   `json:"object"`
	Created        int64    `json:"created"`
	Model          string   `json:"model"`
	Choices        []Choice `json:"choices"`
	Usage          Usage    `json:"usage"`
	ConversationID string   `json:"conversation_id,omitempty"`
}

// Choice represents a single completion choice.
type Choice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message"`
	FinishReason string      `json:"finish_reason"`
	LogProbs     interface{} `json:"logprobs"`
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

// ─── Request Options ────────────────────────────────────────────────────────

// RequestOptions holds per-request header options for cognitive features.
type RequestOptions struct {
	// UserID sets the X-User-ID header for user profiling.
	UserID string
	// ConversationID sets the X-Conversation-ID header to continue a conversation.
	ConversationID string
	// CognitiveDisabled sets X-Cognitive-Disabled to skip cognitive processing.
	CognitiveDisabled bool
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
// Input can be a single string or a slice of strings.
type EmbeddingRequest struct {
	Input      interface{} `json:"input"`
	Model      string      `json:"model"`
	Dimensions *int        `json:"dimensions,omitempty"`
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

// ─── Conversations ──────────────────────────────────────────────────────────

// Conversation represents a conversation summary.
type Conversation struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id,omitempty"`
	Title     string `json:"title"`
	Model     string `json:"model"`
	TurnCount int    `json:"turn_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ConversationTurn represents a single turn in a conversation.
type ConversationTurn struct {
	ID               int    `json:"id"`
	Role             string `json:"role"`
	Content          string `json:"content"`
	Model            string `json:"model,omitempty"`
	PromptTokens     int    `json:"prompt_tokens,omitempty"`
	CompletionTokens int    `json:"completion_tokens,omitempty"`
	LatencyMs        int64  `json:"latency_ms,omitempty"`
	CreatedAt        string `json:"created_at"`
}

// ConversationListResponse represents the response from listing conversations.
type ConversationListResponse struct {
	Object string         `json:"object"`
	Data   []Conversation `json:"data"`
	Total  int            `json:"total"`
}

// ConversationDetail represents a conversation with its turns.
type ConversationDetail struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id,omitempty"`
	Title     string             `json:"title"`
	Model     string             `json:"model"`
	TurnCount int                `json:"turn_count"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
	Turns     []ConversationTurn `json:"turns"`
}

// ConversationSearchResult represents a single search result from conversation search.
type ConversationSearchResult struct {
	ConversationID string `json:"conversation_id"`
	TurnID         int    `json:"turn_id"`
	Role           string `json:"role"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}

// ConversationSearchResponse represents the response from searching conversations.
type ConversationSearchResponse struct {
	Object string                     `json:"object"`
	Data   []ConversationSearchResult `json:"data"`
	Query  string                     `json:"query"`
	Total  int                        `json:"total"`
}

// ConversationDeleteResponse represents the response from deleting a conversation.
type ConversationDeleteResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// ─── Feedback ───────────────────────────────────────────────────────────────

// FeedbackRequest represents a feedback submission request.
type FeedbackRequest struct {
	ConversationID string   `json:"conversation_id"`
	TurnID         int      `json:"turn_id,omitempty"`
	Rating         int      `json:"rating"`
	Comment        string   `json:"comment,omitempty"`
	Correction     string   `json:"correction,omitempty"`
	ChunkIDs       []string `json:"chunk_ids,omitempty"`
}

// FeedbackResponse represents the response from submitting feedback.
type FeedbackResponse struct {
	ID             int    `json:"id"`
	ConversationID string `json:"conversation_id"`
	TurnID         int    `json:"turn_id"`
	Rating         int    `json:"rating"`
	CreatedAt      string `json:"created_at"`
}

// ─── User Profiles ──────────────────────────────────────────────────────────

// UserProfile represents a user profile with expertise and preferences.
type UserProfile struct {
	UserID       string             `json:"user_id"`
	DisplayName  string             `json:"display_name,omitempty"`
	Preferences  map[string]string  `json:"preferences,omitempty"`
	Expertise    map[string]float64 `json:"expertise,omitempty"`
	Topics       map[string]int     `json:"topics,omitempty"`
	TotalQueries int                `json:"total_queries"`
	LastActiveAt string             `json:"last_active_at,omitempty"`
	CreatedAt    string             `json:"created_at,omitempty"`
}

// ProfileUpdateRequest represents a request to update a user profile.
type ProfileUpdateRequest struct {
	DisplayName string            `json:"display_name,omitempty"`
	Preferences map[string]string `json:"preferences,omitempty"`
}

// ─── Knowledge Graph ────────────────────────────────────────────────────────

// KnowledgeNode represents a node in the knowledge graph.
type KnowledgeNode struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	HitCount int    `json:"hit_count"`
}

// KnowledgeEdge represents an edge in the knowledge graph.
type KnowledgeEdge struct {
	ID       int     `json:"id"`
	FromID   string  `json:"from_id"`
	ToID     string  `json:"to_id"`
	Relation string  `json:"relation"`
	Weight   float64 `json:"weight"`
}

// KnowledgeGraphResponse represents the response from querying the knowledge graph.
type KnowledgeGraphResponse struct {
	Object string          `json:"object"`
	Data   []KnowledgeNode `json:"data"`
	Edges  []KnowledgeEdge `json:"edges"`
	Query  string          `json:"query"`
	Total  int             `json:"total"`
}

// ─── Learned Facts ──────────────────────────────────────────────────────────

// Fact represents a learned fact in the knowledge base.
type Fact struct {
	ID             int     `json:"id"`
	Content        string  `json:"content"`
	Source         string  `json:"source"`
	ConversationID string  `json:"conversation_id,omitempty"`
	Confidence     float64 `json:"confidence"`
	Verified       bool    `json:"verified"`
	UsedCount      int     `json:"used_count"`
	CreatedAt      string  `json:"created_at"`
}

// FactCreateRequest represents a request to create a single fact.
type FactCreateRequest struct {
	Content    string  `json:"content"`
	Source     string  `json:"source,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	Verified   bool    `json:"verified,omitempty"`
}

// FactBatchCreateRequest represents a request to create multiple facts.
type FactBatchCreateRequest struct {
	Facts []FactCreateRequest `json:"facts"`
}

// FactUpdateRequest represents a request to update a fact.
type FactUpdateRequest struct {
	Content    *string  `json:"content,omitempty"`
	Confidence *float64 `json:"confidence,omitempty"`
	Verified   *bool    `json:"verified,omitempty"`
}

// FactListResponse represents the response from listing facts.
type FactListResponse struct {
	Object string `json:"object"`
	Data   []Fact `json:"data"`
	Total  int    `json:"total"`
}

// ─── Cognitive Intelligence ─────────────────────────────────────────────────

// CognitiveStatsResponse represents system-wide cognitive statistics.
type CognitiveStatsResponse struct {
	TotalConversations  int     `json:"total_conversations"`
	TotalTurns          int     `json:"total_turns"`
	TotalFeedback       int     `json:"total_feedback"`
	PositiveFeedback    int     `json:"positive_feedback"`
	NegativeFeedback    int     `json:"negative_feedback"`
	TotalUsers          int     `json:"total_users"`
	TotalKnowledgeNodes int     `json:"total_knowledge_nodes"`
	TotalKnowledgeEdges int     `json:"total_knowledge_edges"`
	TotalLearnedFacts   int     `json:"total_learned_facts"`
	VerifiedFacts       int     `json:"verified_facts"`
	AvgFactConfidence   float64 `json:"avg_fact_confidence"`
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
