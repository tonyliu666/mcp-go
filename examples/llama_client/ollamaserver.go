package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// LlamaServer orchestrates the interaction between Ollama and MCP
type LlamaServer struct {
	HTTPServer *server.StreamableHTTPServer
	Ollama     *OllamaClient
	MCPClient  *MCPClientWrapper
}

// NewLlamaServer creates a new orchestrator
func NewLlamaServer(ollamaURL, modelName string) *LlamaServer {
	mcpServer := NewMCPServer()
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	return &LlamaServer{
		HTTPServer: httpServer,
		Ollama:     NewOllamaClient(ollamaURL, modelName),
	}
}

// Start starts the MCP server and initializes the MCP client
func (s *LlamaServer) Start(ctx context.Context) error {
	go func() {
		log.Printf("Starting local MCP server on :3000...")
		if err := s.HTTPServer.Start("127.0.0.1:3000"); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(500 * time.Millisecond)

	mcpClient, err := NewMCPClient(ctx, "http://127.0.0.1:3000/mcp")
	if err != nil {
		return err
	}
	s.MCPClient = mcpClient

	return nil
}

// RunREPL starts the interactive chat loop
func (s *LlamaServer) RunREPL(ctx context.Context) {
	fmt.Printf("\n--- Llama MCP REPL (%s) ---\n", s.Ollama.Model)
	fmt.Println("Type 'exit' to quit.")

	scanner := bufio.NewScanner(os.Stdin)
	history := []ChatMessage{}

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		if strings.ToLower(input) == "exit" {
			break
		}

		history = append(history, ChatMessage{Role: "user", Content: input})

		// Loop for potential multiple tool call rounds
		stepCount := 0
		maxSteps := 5

		for {
			stepCount++
			if stepCount > maxSteps {
				fmt.Println("\n[Error: Maximum tool call steps reached]")
				break
			}

			// Get current tools from MCP server
			mcpTools, err := s.MCPClient.ListTools(ctx)
			if err != nil {
				log.Printf("Error listing tools: %v", err)
				break
			}

			// Send to Ollama
			resp, err := s.Ollama.Chat(ctx, history, mcpTools)
			if err != nil {
				log.Printf("Ollama error: %v", err)
				break
			}

			msg := resp.Message

			if len(msg.ToolCalls) == 0 {
				fmt.Printf("\nLlama: %s\n", msg.Content)
				break
			}

			// Handle tool calls
			toolResults, err := s.MCPClient.HandleToolCalls(ctx, msg.ToolCalls)
			if err != nil {
				log.Printf("Tool handling error: %v", err)
				break
			}
			history = append(history, toolResults...)
		}
	}
}

// Close cleans up resources
func (s *LlamaServer) Close() {
	if s.MCPClient != nil {
		s.MCPClient.Close()
	}
}
