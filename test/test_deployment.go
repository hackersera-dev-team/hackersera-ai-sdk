package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	sdk "github.com/hackersera-dev-team/hackersera-ai-sdk"
)

func main() {
	// Using the deployed endpoint
	baseURL := os.Getenv("HACKERSERA_API_URL")
	if baseURL == "" {
		baseURL = "https://api-ai.hackersera.com"
	}
	apiKey := os.Getenv("HACKERSERA_API_KEY")
	if apiKey == "" {
		log.Fatal("HACKERSERA_API_KEY environment variable is required")
	}

	client := sdk.NewClient(baseURL, apiKey)

	ctx := context.Background()

	fmt.Println("Testing HackersEra AI Deployment")
	fmt.Println("=================================================")
	fmt.Printf("Endpoint: %s\n\n", baseURL)

	// Test 1: Health Check
	fmt.Println("1. Testing Health Endpoint...")
	health, err := client.Health(ctx)
	if err != nil {
		log.Printf("   FAIL Health check: %v\n", err)
	} else {
		fmt.Printf("   OK Health: %s, Version: %s\n\n", health.Status, health.Version)
	}

	// Test 2: Readiness
	fmt.Println("2. Testing Readiness...")
	ready, err := client.Ready(ctx)
	if err != nil {
		log.Printf("   FAIL Readiness: %v\n", err)
	} else {
		fmt.Printf("   OK Ready: %v\n", ready.Ready)
		for k, v := range ready.Checks {
			fmt.Printf("      %s: %s\n", k, v)
		}
		fmt.Println()
	}

	// Test 3: List Models
	fmt.Println("3. Testing List Models...")
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Printf("   FAIL List models: %v\n", err)
	} else {
		fmt.Printf("   OK Found %d models:\n", len(models.Data))
		for _, m := range models.Data {
			fmt.Printf("      - %s (owned by: %s)\n", m.ID, m.OwnedBy)
		}
		fmt.Println()
	}

	// Test 4: Get Specific Model
	fmt.Println("4. Testing Get Model...")
	model, err := client.GetModel(ctx, sdk.ModelDefault)
	if err != nil {
		log.Printf("   FAIL Get model: %v\n", err)
	} else {
		fmt.Printf("   OK Model: %s\n\n", model.ID)
	}

	// Test 5: Chat Completion (with user identity)
	fmt.Println("5. Testing Chat Completion (with user identity)...")
	startTime := time.Now()
	resp, err := client.ChatCompletionWithOptions(ctx, sdk.ChatRequest{
		Model: sdk.ModelDefault,
		Messages: []sdk.Message{
			{Role: "user", Content: "Say 'Hello from HackersEra AI!' and nothing else."},
		},
		User: "test-user",
	}, sdk.RequestOptions{
		UserID: "test-user",
	})
	duration := time.Since(startTime)

	var conversationID string
	if err != nil {
		log.Printf("   FAIL Chat completion: %v\n", err)
	} else {
		fmt.Printf("   OK Response received in %v:\n", duration)
		fmt.Printf("      Content: %s\n", resp.Choices[0].Message.Content)
		fmt.Printf("      Model: %s\n", resp.Model)
		fmt.Printf("      Conversation ID: %s\n", resp.ConversationID)
		fmt.Printf("      Tokens - Prompt: %d, Completion: %d, Total: %d\n\n",
			resp.Usage.PromptTokens,
			resp.Usage.CompletionTokens,
			resp.Usage.TotalTokens)
		conversationID = resp.ConversationID
	}

	// Test 6: Streaming
	fmt.Println("6. Testing Streaming...")
	chunks, errs := client.ChatCompletionStream(ctx, sdk.ChatRequest{
		Model: sdk.ModelDefault,
		Messages: []sdk.Message{
			{Role: "user", Content: "Count from 1 to 3, one number per line."},
		},
	})

	fmt.Print("      Stream output: ")
	streamSuccess := false
	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				streamSuccess = true
				fmt.Println()
				goto streamDone
			}
			if len(chunk.Choices) > 0 {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		case err, ok := <-errs:
			if ok && err != nil {
				log.Printf("\n   FAIL Streaming: %v\n", err)
				goto streamDone
			}
		}
	}

streamDone:
	if streamSuccess {
		fmt.Println("   OK Streaming completed successfully\n")
	}

	// Test 7: Embeddings
	fmt.Println("7. Testing Embeddings...")
	emb, err := client.CreateEmbedding(ctx, sdk.EmbeddingRequest{
		Input: "Hello world",
		Model: sdk.ModelEmbedding,
	})
	if err != nil {
		log.Printf("   FAIL Embeddings: %v\n", err)
	} else {
		fmt.Printf("   OK Embedding created:\n")
		fmt.Printf("      Dimensions: %d\n", len(emb.Data[0].Embedding))
		fmt.Printf("      Total tokens: %d\n\n", emb.Usage.TotalTokens)
	}

	// Test 8: Upload Document (RAG)
	fmt.Println("8. Testing Document Upload...")
	doc, err := client.UploadDocument(ctx, sdk.DocumentUploadRequest{
		Content:  "HackersEra is an Indian cybersecurity company founded in 2018 that provides AI-powered penetration testing and vulnerability assessment services.",
		Filename: "test-doc.txt",
		Tags:     map[string]string{"test": "deployment"},
	})
	if err != nil {
		log.Printf("   FAIL Document upload: %v\n", err)
	} else {
		fmt.Printf("   OK Document uploaded: %s (status: %s)\n", doc.ID, doc.Status)

		// Poll for indexing
		fmt.Print("      Waiting for indexing...")
		for i := 0; i < 10; i++ {
			time.Sleep(500 * time.Millisecond)
			d, err := client.GetDocument(ctx, doc.ID)
			if err == nil && d.Status == "indexed" {
				fmt.Printf(" indexed (%d chunks)\n", d.ChunkCount)
				break
			}
			if err == nil && d.Status == "failed" {
				fmt.Printf(" failed: %s\n", d.Error)
				break
			}
			fmt.Print(".")
		}
		fmt.Println()
	}

	// Test 9: List Documents
	fmt.Println("9. Testing List Documents...")
	docs, err := client.ListDocuments(ctx)
	if err != nil {
		log.Printf("   FAIL List documents: %v\n", err)
	} else {
		fmt.Printf("   OK Found %d documents\n\n", docs.Total)
	}

	// Test 10: Search
	fmt.Println("10. Testing Search...")
	searchResults, err := client.Search(ctx, sdk.SearchRequest{
		Query: "cybersecurity penetration testing",
	})
	if err != nil {
		log.Printf("   FAIL Search: %v\n", err)
	} else {
		fmt.Printf("   OK Found %d results for %q\n", searchResults.Total, searchResults.Query)
		for _, r := range searchResults.Data {
			fmt.Printf("      [%s] score=%.2f\n", r.Filename, r.Score)
		}
		fmt.Println()
	}

	// Test 11: Submit Feedback
	fmt.Println("11. Testing Feedback...")
	if conversationID != "" {
		fb, err := client.SubmitFeedback(ctx, sdk.FeedbackRequest{
			ConversationID: conversationID,
			Rating:         1,
			Comment:        "Deployment test feedback",
		})
		if err != nil {
			log.Printf("   FAIL Feedback: %v\n", err)
		} else {
			fmt.Printf("   OK Feedback ID: %d, Rating: %d\n\n", fb.ID, fb.Rating)
		}
	} else {
		fmt.Println("   SKIP No conversation ID available\n")
	}

	// Test 12: List Conversations
	fmt.Println("12. Testing List Conversations...")
	convos, err := client.ListConversations(ctx, 5)
	if err != nil {
		log.Printf("   FAIL List conversations: %v\n", err)
	} else {
		fmt.Printf("   OK Found %d conversations\n", convos.Total)
		for _, conv := range convos.Data {
			fmt.Printf("      [%s] %s (%d turns)\n", conv.ID, conv.Title, conv.TurnCount)
		}
		fmt.Println()
	}

	// Test 13: Get Conversation Detail
	fmt.Println("13. Testing Get Conversation...")
	if conversationID != "" {
		detail, err := client.GetConversation(ctx, conversationID)
		if err != nil {
			log.Printf("   FAIL Get conversation: %v\n", err)
		} else {
			fmt.Printf("   OK Conversation: %s, Turns: %d\n", detail.Title, detail.TurnCount)
			for _, turn := range detail.Turns {
				preview := turn.Content
				if len(preview) > 60 {
					preview = preview[:60] + "..."
				}
				fmt.Printf("      [%s] %s\n", turn.Role, preview)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("   SKIP No conversation ID available\n")
	}

	// Test 14: Search Conversations
	fmt.Println("14. Testing Search Conversations...")
	convSearch, err := client.SearchConversations(ctx, "Hello", 5)
	if err != nil {
		log.Printf("   FAIL Search conversations: %v\n", err)
	} else {
		fmt.Printf("   OK Found %d results for %q\n\n", convSearch.Total, convSearch.Query)
	}

	// Test 15: User Profile — Get
	fmt.Println("15. Testing Get Profile...")
	profile, err := client.GetProfile(ctx, "test-user")
	if err != nil {
		log.Printf("   FAIL Get profile: %v\n", err)
	} else {
		fmt.Printf("   OK User: %s, Queries: %d\n\n", profile.UserID, profile.TotalQueries)
	}

	// Test 16: User Profile — Update
	fmt.Println("16. Testing Update Profile...")
	updatedProfile, err := client.UpdateProfile(ctx, "test-user", sdk.ProfileUpdateRequest{
		DisplayName: "Test User",
		Preferences: map[string]string{"language": "go", "detail_level": "concise"},
	})
	if err != nil {
		log.Printf("   FAIL Update profile: %v\n", err)
	} else {
		fmt.Printf("   OK Updated: %s (display: %s)\n\n", updatedProfile.UserID, updatedProfile.DisplayName)
	}

	// Test 17: Create Fact
	fmt.Println("17. Testing Create Fact...")
	fact, err := client.CreateFact(ctx, sdk.FactCreateRequest{
		Content:    "HackersEra SDK supports Go, Python, and Node.js",
		Source:     "manual",
		Confidence: 0.9,
		Verified:   true,
	})
	if err != nil {
		log.Printf("   FAIL Create fact: %v\n", err)
	} else {
		fmt.Printf("   OK Fact ID: %d, Confidence: %.2f\n\n", fact.ID, fact.Confidence)
	}

	// Test 18: Batch Create Facts
	fmt.Println("18. Testing Batch Create Facts...")
	batchFacts, err := client.CreateFacts(ctx, []sdk.FactCreateRequest{
		{Content: "Go SDK uses zero external dependencies", Source: "docs", Confidence: 0.95},
		{Content: "API is OpenAI-compatible", Source: "docs", Confidence: 0.99, Verified: true},
	})
	if err != nil {
		log.Printf("   FAIL Batch create facts: %v\n", err)
	} else {
		fmt.Printf("   OK Created %d facts\n\n", batchFacts.Total)
	}

	// Test 19: List Facts
	fmt.Println("19. Testing List Facts...")
	facts, err := client.ListFacts(ctx, 10, nil)
	if err != nil {
		log.Printf("   FAIL List facts: %v\n", err)
	} else {
		fmt.Printf("   OK Found %d facts\n", facts.Total)
		for _, f := range facts.Data {
			preview := f.Content
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			fmt.Printf("      [%d] %s (conf: %.2f)\n", f.ID, preview, f.Confidence)
		}
		fmt.Println()
	}

	// Test 20: Update Fact
	if fact != nil {
		fmt.Println("20. Testing Update Fact...")
		updated, err := client.UpdateFact(ctx, fact.ID, sdk.FactUpdateRequest{
			Verified:   sdk.BoolPtr(true),
			Confidence: sdk.Float64Ptr(0.99),
		})
		if err != nil {
			log.Printf("   FAIL Update fact: %v\n", err)
		} else {
			fmt.Printf("   OK Updated fact %d: confidence=%.2f, verified=%v\n\n", updated.ID, updated.Confidence, updated.Verified)
		}
	}

	// Test 21: Knowledge Graph
	fmt.Println("21. Testing Knowledge Graph...")
	graph, err := client.QueryKnowledgeGraph(ctx, "cybersecurity", 10)
	if err != nil {
		log.Printf("   FAIL Knowledge graph: %v\n", err)
	} else {
		fmt.Printf("   OK Found %d nodes, %d edges for %q\n", graph.Total, len(graph.Edges), graph.Query)
		for _, node := range graph.Data {
			fmt.Printf("      [%s] %s (hits: %d)\n", node.ID, node.Label, node.HitCount)
		}
		fmt.Println()
	}

	// Test 22: Cognitive Stats
	fmt.Println("22. Testing Cognitive Stats...")
	cogStats, err := client.GetCognitiveStats(ctx)
	if err != nil {
		log.Printf("   FAIL Cognitive stats: %v\n", err)
	} else {
		fmt.Printf("   OK Conversations: %d, Turns: %d, Facts: %d\n", cogStats.TotalConversations, cogStats.TotalTurns, cogStats.TotalLearnedFacts)
		fmt.Printf("      Graph: %d nodes, %d edges\n", cogStats.TotalKnowledgeNodes, cogStats.TotalKnowledgeEdges)
		fmt.Printf("      Feedback: %d (+%d / -%d)\n\n", cogStats.TotalFeedback, cogStats.PositiveFeedback, cogStats.NegativeFeedback)
	}

	// Test 23: Usage Stats
	fmt.Println("23. Testing Usage Stats...")
	usage, err := client.GetUsage(ctx)
	if err != nil {
		log.Printf("   FAIL Usage: %v\n", err)
	} else {
		fmt.Printf("   OK Total requests: %d, Total tokens: %d\n\n", usage.TotalRequests, usage.TotalTokens)
	}

	// Test 24: Recent Usage
	fmt.Println("24. Testing Recent Usage...")
	recent, err := client.GetRecentUsage(ctx)
	if err != nil {
		log.Printf("   FAIL Recent usage: %v\n", err)
	} else {
		fmt.Printf("   OK Found %d recent records\n\n", recent.Count)
	}

	// Test 25: Cache Stats
	fmt.Println("25. Testing Cache Stats...")
	cacheStats, err := client.GetCacheStats(ctx)
	if err != nil {
		log.Printf("   FAIL Cache stats: %v\n", err)
	} else {
		fmt.Printf("   OK Entries: %d, Hits: %d, Tokens saved: %d\n\n",
			cacheStats.ActiveEntries, cacheStats.TotalHits, cacheStats.TokensSaved)
	}

	// Test 26: Metrics
	fmt.Println("26. Testing Metrics...")
	metrics, err := client.GetMetrics(ctx)
	if err != nil {
		log.Printf("   FAIL Metrics: %v\n", err)
	} else {
		preview := metrics
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("   OK Metrics received (%d bytes):\n      %s\n\n", len(metrics), preview)
	}

	// Test 27: Cleanup — delete test document
	if doc != nil {
		fmt.Println("27. Testing Document Delete...")
		del, err := client.DeleteDocument(ctx, doc.ID)
		if err != nil {
			log.Printf("   FAIL Delete: %v\n", err)
		} else {
			fmt.Printf("   OK Deleted: %s = %v\n\n", del.ID, del.Deleted)
		}
	}

	// Test 28: Cleanup — delete test conversation
	if conversationID != "" {
		fmt.Println("28. Testing Conversation Delete...")
		delConv, err := client.DeleteConversation(ctx, conversationID)
		if err != nil {
			log.Printf("   FAIL Delete conversation: %v\n", err)
		} else {
			fmt.Printf("   OK Deleted: %s = %v\n\n", delConv.ID, delConv.Deleted)
		}
	}

	fmt.Println("=================================================")
	fmt.Println("Deployment testing completed!")
}
