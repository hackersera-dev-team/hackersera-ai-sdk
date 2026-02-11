package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
	fmt.Printf("Status: %s, CLI: %s\n\n", health.Status, health.ClaudeCLI)

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

	// ─── Chat Completion ─────────────────────────────────────────────────
	fmt.Println("=== Chat Completion ===")
	resp, err := client.ChatCompletion(ctx, sdk.ChatRequest{
		Model: "sonnet",
		Messages: []sdk.Message{
			{Role: "system", Content: "You are a helpful assistant. Be concise."},
			{Role: "user", Content: "What is Go programming language?"},
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
		Model: "sonnet",
		Messages: []sdk.Message{
			{Role: "user", Content: "Write a haiku about coding"},
		},
	})

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				fmt.Println()
				return
			}
			if len(chunk.Choices) > 0 {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		case err, ok := <-errs:
			if ok && err != nil {
				log.Fatalf("\nStream error: %v", err)
			}
			fmt.Println()
			return
		}
	}
}
