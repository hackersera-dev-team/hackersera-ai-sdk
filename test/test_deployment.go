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
		baseURL = "http://hackersera-ai.cloudjiffy.net"
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

	// Test 5: Chat Completion
	fmt.Println("5. Testing Chat Completion...")
	startTime := time.Now()
	resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
		Model: sdk.ModelDefault,
		Messages: []sdk.Message{
			{Role: "user", Content: "Say 'Hello from HackersEra AI!' and nothing else."},
		},
	})
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("   FAIL Chat completion: %v\n", err)
	} else {
		fmt.Printf("   OK Response received in %v:\n", duration)
		fmt.Printf("      Content: %s\n", resp.Choices[0].Message.Content)
		fmt.Printf("      Model: %s\n", resp.Model)
		fmt.Printf("      Tokens - Prompt: %d, Completion: %d, Total: %d\n\n",
			resp.Usage.PromptTokens,
			resp.Usage.CompletionTokens,
			resp.Usage.TotalTokens)
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

	// Test 11: Usage Stats
	fmt.Println("11. Testing Usage Stats...")
	usage, err := client.GetUsage(ctx)
	if err != nil {
		log.Printf("   FAIL Usage: %v\n", err)
	} else {
		fmt.Printf("   OK Total requests: %d, Total tokens: %d\n\n", usage.TotalRequests, usage.TotalTokens)
	}

	// Test 12: Cache Stats
	fmt.Println("12. Testing Cache Stats...")
	cacheStats, err := client.GetCacheStats(ctx)
	if err != nil {
		log.Printf("   FAIL Cache stats: %v\n", err)
	} else {
		fmt.Printf("   OK Entries: %d, Hits: %d, Tokens saved: %d\n\n",
			cacheStats.ActiveEntries, cacheStats.TotalHits, cacheStats.TokensSaved)
	}

	// Test 13: Cleanup — delete test document
	if doc != nil {
		fmt.Println("13. Testing Document Delete...")
		del, err := client.DeleteDocument(ctx, doc.ID)
		if err != nil {
			log.Printf("   FAIL Delete: %v\n", err)
		} else {
			fmt.Printf("   OK Deleted: %s = %v\n\n", del.ID, del.Deleted)
		}
	}

	fmt.Println("=================================================")
	fmt.Println("Deployment testing completed!")
}
