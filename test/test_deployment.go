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
		log.Fatal("❌ HACKERSERA_API_KEY environment variable is required")
	}

	client := sdk.NewClient(baseURL, apiKey)

	ctx := context.Background()

	fmt.Println("🧪 Testing HackersEra AI Deployment")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Printf("Endpoint: %s\n\n", baseURL)

	// Test 1: Health Check
	fmt.Println("1️⃣  Testing Health Endpoint...")
	health, err := client.Health(ctx)
	if err != nil {
		log.Printf("❌ Health check failed: %v\n", err)
	} else {
		fmt.Printf("✅ Health: %s, Version: %s\n\n", health.Status, health.Version)
	}

	// Test 2: List Models
	fmt.Println("2️⃣  Testing List Models...")
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Printf("❌ List models failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d models:\n", len(models.Data))
		for _, m := range models.Data {
			fmt.Printf("   - %s (owned by: %s)\n", m.ID, m.OwnedBy)
		}
		fmt.Println()
	}

	// Test 3: Get Specific Model
	fmt.Println("3️⃣  Testing Get Model...")
	model, err := client.GetModel(ctx, sdk.ModelDefault)
	if err != nil {
		log.Printf("❌ Get model failed: %v\n", err)
	} else {
		fmt.Printf("✅ Model: %s\n\n", model.ID)
	}

	// Test 4: Chat Completion
	fmt.Println("4️⃣  Testing Chat Completion...")
	startTime := time.Now()
	resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
		Model: sdk.ModelDefault,
		Messages: []sdk.Message{
			{Role: "user", Content: "Say 'Hello from HackersEra AI!' and nothing else."},
		},
	})
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ Chat completion failed: %v\n", err)
	} else {
		fmt.Printf("✅ Response received in %v:\n", duration)
		fmt.Printf("   Content: %s\n", resp.Choices[0].Message.Content)
		fmt.Printf("   Model: %s\n", resp.Model)
		fmt.Printf("   Tokens - Prompt: %d, Completion: %d, Total: %d\n\n",
			resp.Usage.PromptTokens,
			resp.Usage.CompletionTokens,
			resp.Usage.TotalTokens)
	}

	// Test 5: Streaming
	fmt.Println("5️⃣  Testing Streaming...")
	chunks, errs := client.ChatCompletionStream(ctx, sdk.ChatRequest{
		Model: sdk.ModelDefault,
		Messages: []sdk.Message{
			{Role: "user", Content: "Count from 1 to 3, one number per line."},
		},
	})

	fmt.Print("   Stream output: ")
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
				log.Printf("\n❌ Streaming failed: %v\n", err)
				goto streamDone
			}
		}
	}

streamDone:
	if streamSuccess {
		fmt.Println("✅ Streaming completed successfully\n")
	}

	// Test 6: Embeddings (if available)
	fmt.Println("6️⃣  Testing Embeddings...")
	emb, err := client.CreateEmbedding(ctx, sdk.EmbeddingRequest{
		Input: "Hello world",
		Model: sdk.ModelEmbedding,
	})
	if err != nil {
		log.Printf("❌ Embeddings failed: %v (this might not be implemented)\n", err)
	} else {
		fmt.Printf("✅ Embedding created:\n")
		fmt.Printf("   Dimensions: %d\n", len(emb.Data[0].Embedding))
		fmt.Printf("   Total tokens: %d\n\n", emb.Usage.TotalTokens)
	}

	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println("✅ Deployment testing completed!")
}
