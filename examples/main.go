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
		baseURL = "http://localhost:8080"
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
			fmt.Printf("  [%s] %s (score: %.2f)\n", r.Filename, r.Content[:80]+"...", r.Score)
		}
		fmt.Println()
	}

	// ─── Chat Completion (with RAG context) ──────────────────────────────
	fmt.Println("=== Chat Completion (RAG-augmented) ===")
	resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
		Model: sdk.ModelDefault,
		Messages: []sdk.Message{
			{Role: "user", Content: "What is HackersEra?"},
		},
	})
	if err != nil {
		log.Fatalf("Chat completion failed: %v", err)
	}
	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
	fmt.Printf("Tokens: %d prompt + %d completion = %d total\n\n",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)

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
				fmt.Println("\n")
				goto done
			}
			if len(chunk.Choices) > 0 {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		case err, ok := <-errs:
			if ok && err != nil {
				log.Fatalf("\nStream error: %v", err)
			}
			fmt.Println("\n")
			goto done
		}
	}

done:
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
}
