package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

// OllamaClient handles communication with a local Ollama instance
type OllamaClient struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

// NewOllamaClient creates a new client
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.1"
	}
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		Client:  &http.Client{},
	}
}

// ChatRequest represents the request body for Ollama chat API
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Tools    []Tool        `json:"tools,omitempty"`
	Stream   bool          `json:"stream"`
}

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Tool represents a tool definition for Ollama
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction describes the function to call
type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ToolCall represents a request from the model to call a tool
type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ChatResponse represents the response from Ollama
type ChatResponse struct {
	Model     string      `json:"model"`
	CreatedAt string      `json:"created_at"`
	Message   ChatMessage `json:"message"`
	Done      bool        `json:"done"`
}

// Chat sends a message to Ollama and returns the response
func (c *OllamaClient) Chat(ctx context.Context, messages []ChatMessage, tools []mcp.Tool) (*ChatResponse, error) {
	// Convert MCP tools to Ollama tools
	ollamaTools := make([]Tool, len(tools))
	for i, t := range tools {
		// MCP tools schema is already compatible with OpenAI/Ollama function calling schema
		// But we need to ensure it's wrapped correctly

		// The inputSchema in mcp.Tool is technically interface{}, usually a map
		schemaBytes, err := json.Marshal(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool schema: %w", err)
		}

		ollamaTools[i] = Tool{
			Type: "function",
			Function: ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  json.RawMessage(schemaBytes),
			},
		}
	}

	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Tools:    ollamaTools,
		Stream:   false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/chat", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the body into a buffer so we can log it AND decode it
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}
