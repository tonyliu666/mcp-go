package main

import (
	"context"
	"fmt"
	"log"
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

// Ask runs the Ollama+MCP agentic loop for a single question and returns
// the final text response. It is safe to call concurrently.
func (s *LlamaServer) Ask(ctx context.Context, question string) (string, error) {
	history := []ChatMessage{
		{Role: "user", Content: question},
	}

	const maxSteps = 5
	for step := 0; step < maxSteps; step++ {
		// Get current tools from MCP server
		mcpTools, err := s.MCPClient.ListTools(ctx)
		if err != nil {
			return "", fmt.Errorf("listing tools: %w", err)
		}

		// Send to Ollama
		resp, err := s.Ollama.Chat(ctx, history, mcpTools)
		if err != nil {
			return "", fmt.Errorf("ollama chat: %w", err)
		}

		msg := resp.Message

		// No tool calls → Ollama produced a final answer
		if len(msg.ToolCalls) == 0 {
			return msg.Content, nil
		}

		// Execute tool calls and feed results back into history
		toolResults, err := s.MCPClient.HandleToolCalls(ctx, msg.ToolCalls)
		if err != nil {
			return "", fmt.Errorf("handling tool calls: %w", err)
		}
		history = append(history, toolResults...)
	}

	return "", fmt.Errorf("maximum tool call steps reached without a final answer")
}

// Close cleans up resources
func (s *LlamaServer) Close() {
	if s.MCPClient != nil {
		s.MCPClient.Close()
	}
}
