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
	baseURL := os.Getenv("HACKERSERA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api-ai.hackersera.com"
	}
	apiKey := os.Getenv("HACKERSERA_API_KEY")

	client := sdk.NewClient(baseURL, apiKey)
	ctx := context.Background()

	// ─── Health Check ────────────────────────────────────────────────────
	fmt.Println("=== Health Check ===")
	health, err := client.Health(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Printf("Status: %s, Version: %s\n\n", health.Status, health.Version)

	// ─── List Models ─────────────────────────────────────────────────────
	fmt.Println("=== Available Models ===")
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Fatalf("List models failed: %v", err)
	}
	for _, m := range models.Data {
		fmt.Printf("  %s (owned by: %s)\n", m.ID, m.OwnedBy)
	}
	fmt.Println()

	// ─── Upload Document (RAG) ───────────────────────────────────────────
	fmt.Println("=== Upload Document ===")
	doc, err := client.UploadDocument(ctx, sdk.DocumentUploadRequest{
		Content:  "HackersEra is an Indian cybersecurity company founded in 2018. They specialize in AI-powered security solutions including penetration testing, red teaming, and vulnerability assessment.",
		Filename: "about-hackersera.txt",
		Tags:     map[string]string{"category": "company"},
	})
	if err != nil {
		log.Fatalf("Upload document failed: %v", err)
	}
	fmt.Printf("Document ID: %s, Status: %s\n", doc.ID, doc.Status)

	// Wait for indexing
	fmt.Print("Waiting for indexing...")
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		d, err := client.GetDocument(ctx, doc.ID)
		if err == nil && d.Status == "indexed" {
			fmt.Printf(" done (%d chunks)\n\n", d.ChunkCount)
			break
		}
		if err == nil && d.Status == "failed" {
			fmt.Printf(" failed: %s\n\n", d.Error)
			break
		}
		fmt.Print(".")
	}

	// ─── Search Knowledge Base ───────────────────────────────────────────
	fmt.Println("=== Search ===")
	results, err := client.Search(ctx, sdk.SearchRequest{
		Query: "cybersecurity company",
	})
	if err != nil {
		log.Printf("Search failed: %v\n\n", err)
	} else {
		fmt.Printf("Found %d results for %q:\n", results.Total, results.Query)
		for _, r := range results.Data {
			preview := r.Content
			if len(preview) > 80 {
				preview = preview[:80] + "..."
			}
			fmt.Printf("  [%s] %s (score: %.2f)\n", r.Filename, preview, r.Score)
		}
		fmt.Println()
	}

	// ─── Chat Completion (with RAG context + user identity) ──────────────
	fmt.Println("=== Chat Completion (RAG-augmented, with user) ===")
	resp, err := client.ChatCompletionWithOptions(ctx, sdk.ChatRequest{
		Model: sdk.ModelDefault,
		Messages: []sdk.Message{
			{Role: "user", Content: "What is HackersEra?"},
		},
		User: "demo-user",
	}, sdk.RequestOptions{
		UserID: "demo-user",
	})
	if err != nil {
		log.Fatalf("Chat completion failed: %v", err)
	}
	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
	fmt.Printf("Tokens: %d prompt + %d completion = %d total\n", resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
	fmt.Printf("Conversation ID: %s\n\n", resp.ConversationID)

	conversationID := resp.ConversationID

	// ─── Submit Feedback ─────────────────────────────────────────────────
	if conversationID != "" {
		fmt.Println("=== Submit Feedback ===")
		fb, err := client.SubmitFeedback(ctx, sdk.FeedbackRequest{
			ConversationID: conversationID,
			Rating:         1,
			Comment:        "Clear and accurate explanation",
		})
		if err != nil {
			log.Printf("Feedback failed: %v\n\n", err)
		} else {
			fmt.Printf("Feedback ID: %d, Rating: %d\n\n", fb.ID, fb.Rating)
		}
	}

	// ─── Streaming ───────────────────────────────────────────────────────
	fmt.Println("=== Streaming Chat ===")
	chunks, errs := client.ChatCompletionStream(ctx, sdk.ChatRequest{
		Model: sdk.ModelDefault,
		Messages: []sdk.Message{
			{Role: "user", Content: "Write a haiku about coding"},
		},
	})

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				fmt.Println()
				goto done
			}
			if len(chunk.Choices) > 0 {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		case err, ok := <-errs:
			if ok && err != nil {
				log.Fatalf("\nStream error: %v", err)
			}
			fmt.Println()
			goto done
		}
	}

done:
	// ─── List Conversations ──────────────────────────────────────────────
	fmt.Println("=== Conversations ===")
	convos, err := client.ListConversations(ctx, 5)
	if err != nil {
		log.Printf("List conversations failed: %v\n\n", err)
	} else {
		fmt.Printf("Found %d conversations:\n", convos.Total)
		for _, conv := range convos.Data {
			fmt.Printf("  [%s] %s (%d turns)\n", conv.ID, conv.Title, conv.TurnCount)
		}
		fmt.Println()
	}

	// ─── Get Conversation Detail ─────────────────────────────────────────
	if conversationID != "" {
		fmt.Println("=== Conversation Detail ===")
		detail, err := client.GetConversation(ctx, conversationID)
		if err != nil {
			log.Printf("Get conversation failed: %v\n\n", err)
		} else {
			fmt.Printf("Conversation: %s (%d turns)\n", detail.Title, detail.TurnCount)
			for _, turn := range detail.Turns {
				preview := turn.Content
				if len(preview) > 60 {
					preview = preview[:60] + "..."
				}
				fmt.Printf("  [%s] %s\n", turn.Role, preview)
			}
			fmt.Println()
		}
	}

	// ─── Search Conversations ────────────────────────────────────────────
	fmt.Println("=== Search Conversations ===")
	convSearch, err := client.SearchConversations(ctx, "HackersEra", 5)
	if err != nil {
		log.Printf("Search conversations failed: %v\n\n", err)
	} else {
		fmt.Printf("Found %d results for %q:\n", convSearch.Total, convSearch.Query)
		for _, r := range convSearch.Data {
			preview := r.Content
			if len(preview) > 60 {
				preview = preview[:60] + "..."
			}
			fmt.Printf("  [%s] %s: %s\n", r.ConversationID, r.Role, preview)
		}
		fmt.Println()
	}

	// ─── User Profile ────────────────────────────────────────────────────
	fmt.Println("=== User Profile ===")
	profile, err := client.GetProfile(ctx, "demo-user")
	if err != nil {
		log.Printf("Get profile failed: %v\n\n", err)
	} else {
		fmt.Printf("User: %s, Queries: %d\n", profile.UserID, profile.TotalQueries)
		if len(profile.Expertise) > 0 {
			fmt.Print("  Expertise: ")
			for topic, score := range profile.Expertise {
				fmt.Printf("%s=%.2f ", topic, score)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// ─── Update Profile ──────────────────────────────────────────────────
	fmt.Println("=== Update Profile ===")
	updatedProfile, err := client.UpdateProfile(ctx, "demo-user", sdk.ProfileUpdateRequest{
		DisplayName: "Demo User",
		Preferences: map[string]string{
			"language":     "go",
			"detail_level": "concise",
		},
	})
	if err != nil {
		log.Printf("Update profile failed: %v\n\n", err)
	} else {
		fmt.Printf("Updated: %s (display: %s)\n\n", updatedProfile.UserID, updatedProfile.DisplayName)
	}

	// ─── Create Fact ─────────────────────────────────────────────────────
	fmt.Println("=== Create Fact ===")
	fact, err := client.CreateFact(ctx, sdk.FactCreateRequest{
		Content:    "HackersEra was founded in 2018 in India",
		Source:     "manual",
		Confidence: 0.95,
		Verified:   true,
	})
	if err != nil {
		log.Printf("Create fact failed: %v\n\n", err)
	} else {
		fmt.Printf("Fact ID: %d, Confidence: %.2f, Verified: %v\n\n", fact.ID, fact.Confidence, fact.Verified)
	}

	// ─── List Facts ──────────────────────────────────────────────────────
	fmt.Println("=== List Facts ===")
	facts, err := client.ListFacts(ctx, 10, nil)
	if err != nil {
		log.Printf("List facts failed: %v\n\n", err)
	} else {
		fmt.Printf("Found %d facts:\n", facts.Total)
		for _, f := range facts.Data {
			preview := f.Content
			if len(preview) > 60 {
				preview = preview[:60] + "..."
			}
			fmt.Printf("  [%d] %s (confidence: %.2f, verified: %v)\n", f.ID, preview, f.Confidence, f.Verified)
		}
		fmt.Println()
	}

	// ─── Knowledge Graph ─────────────────────────────────────────────────
	fmt.Println("=== Knowledge Graph ===")
	graph, err := client.QueryKnowledgeGraph(ctx, "cybersecurity", 10)
	if err != nil {
		log.Printf("Knowledge graph failed: %v\n\n", err)
	} else {
		fmt.Printf("Found %d nodes for %q:\n", graph.Total, graph.Query)
		for _, node := range graph.Data {
			fmt.Printf("  [%s] %s (type: %s, hits: %d)\n", node.ID, node.Label, node.Type, node.HitCount)
		}
		if len(graph.Edges) > 0 {
			fmt.Printf("Edges: %d\n", len(graph.Edges))
		}
		fmt.Println()
	}

	// ─── Cognitive Stats ─────────────────────────────────────────────────
	fmt.Println("=== Cognitive Stats ===")
	cogStats, err := client.GetCognitiveStats(ctx)
	if err != nil {
		log.Printf("Cognitive stats failed: %v\n\n", err)
	} else {
		fmt.Printf("Conversations: %d, Turns: %d, Facts: %d (verified: %d)\n",
			cogStats.TotalConversations, cogStats.TotalTurns, cogStats.TotalLearnedFacts, cogStats.VerifiedFacts)
		fmt.Printf("Knowledge Graph: %d nodes, %d edges\n", cogStats.TotalKnowledgeNodes, cogStats.TotalKnowledgeEdges)
		fmt.Printf("Feedback: %d total (%d positive, %d negative)\n\n",
			cogStats.TotalFeedback, cogStats.PositiveFeedback, cogStats.NegativeFeedback)
	}

	// ─── Usage Stats ─────────────────────────────────────────────────────
	fmt.Println("=== Usage Stats ===")
	usage, err := client.GetUsage(ctx)
	if err != nil {
		log.Printf("Usage failed: %v\n", err)
	} else {
		fmt.Printf("Total requests: %d, Total tokens: %d\n\n", usage.TotalRequests, usage.TotalTokens)
	}

	// ─── Cleanup: Delete Document ────────────────────────────────────────
	fmt.Println("=== Cleanup ===")
	del, err := client.DeleteDocument(ctx, doc.ID)
	if err != nil {
		log.Printf("Delete failed: %v\n", err)
	} else {
		fmt.Printf("Deleted document %s: %v\n", del.ID, del.Deleted)
	}

	// Clean up conversation
	if conversationID != "" {
		delConv, err := client.DeleteConversation(ctx, conversationID)
		if err != nil {
			log.Printf("Delete conversation failed: %v\n", err)
		} else {
			fmt.Printf("Deleted conversation %s: %v\n", delConv.ID, delConv.Deleted)
		}
	}
}
