package main

import (
	"context"
	"flag"
	"log"
)

func main() {
	modelName := flag.String("model", "llama3.1", "Ollama model to use")
	ollamaURL := flag.String("ollama-url", "http://127.0.0.1:11434", "Ollama API base URL")
	flag.Parse()

	ctx := context.Background()

	// Initialize the Llama MCP Orchestrator
	orchestrator := NewLlamaServer(*ollamaURL, *modelName)
	defer orchestrator.Close()

	// Start the MCP server and client
	if err := orchestrator.Start(ctx); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	// Run the REPL
	orchestrator.RunREPL(ctx)
}
