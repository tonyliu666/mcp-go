package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPClientWrapper wraps the MCP client and provides helper methods
type MCPClientWrapper struct {
	Client *client.Client
}

// NewMCPClient creates and initializes a new MCP client
func NewMCPClient(ctx context.Context, url string) (*MCPClientWrapper, error) {
	mcpClient, err := client.NewStreamableHttpClient(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}

	if err := mcpClient.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start MCP client: %w", err)
	}

	if _, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{
				Name:    "LlamaClientDemo",
				Version: "1.0.0",
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	return &MCPClientWrapper{Client: mcpClient}, nil
}

// Close closes the MCP client connection
func (w *MCPClientWrapper) Close() error {
	return w.Client.Close()
}

// ListTools returns the list of tools from the MCP server
func (w *MCPClientWrapper) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	resp, err := w.Client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Tools, nil
}

// HandleToolCalls processes tool calls from the model
func (w *MCPClientWrapper) HandleToolCalls(ctx context.Context, toolCalls []ToolCall) ([]ChatMessage, error) {
	var toolResults []ChatMessage

	for _, tc := range toolCalls {
		fmt.Printf("[Calling Tool: %s with %s]\n", tc.Function.Name, tc.Function.Arguments)

		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			log.Printf("Failed to parse tool arguments: %v", err)
			continue
		}

		result, err := w.Client.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      tc.Function.Name,
				Arguments: args,
			},
		})

		var resultText string
		if err != nil {
			resultText = fmt.Sprintf("Error: %v", err)
		} else {
			for _, c := range result.Content {
				switch v := c.(type) {
				case mcp.TextContent:
					resultText += v.Text
				case *mcp.TextContent:
					resultText += v.Text
				}
			}
		}

		fmt.Printf("[Tool Result: %s]\n", resultText)
		toolResults = append(toolResults, ChatMessage{
			Role:    "tool",
			Content: resultText,
		})
	}

	return toolResults, nil
}
