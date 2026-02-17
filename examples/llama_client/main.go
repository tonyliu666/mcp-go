package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	modelName := flag.String("model", "llama3.1", "Ollama model to use")
	ollamaURL := flag.String("ollama-url", "http://127.0.0.1:11434", "Ollama API base URL")
	flag.Parse()

	ctx := context.Background()

	// 1. Create and Start the MCP Server
	mcpServer := NewMCPServer()
	httpServer := server.NewStreamableHTTPServer(mcpServer)

	go func() {
		log.Printf("Starting local MCP server on :3000...")
		if err := httpServer.Start("127.0.0.1:3000"); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(500 * time.Millisecond)

	// 2. Initialize MCP Client
	mcpClient, err := client.NewStreamableHttpClient("http://127.0.0.1:3000/mcp")
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}

	if err := mcpClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start MCP client: %v", err)
	}

	if _, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{
				Name:    "LlamaClientDemo",
				Version: "1.0.0",
			},
		},
	}); err != nil {
		log.Fatalf("Failed to initialize MCP client: %v", err)
	}
	defer mcpClient.Close()

	// 3. Initialize Ollama Client
	ollama := NewOllamaClient(*ollamaURL, *modelName)

	fmt.Printf("\n--- Llama MCP REPL (%s) ---\n", *modelName)
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
			mcpTools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
			if err != nil {
				log.Printf("Error listing tools: %v", err)
				break
			}
			log.Println("history: ", history)
			log.Println("mcpTools: ", mcpTools)

			// Send to Ollama
			resp, err := ollama.Chat(ctx, history, mcpTools.Tools)
			if err != nil {
				log.Printf("Ollama error: %v", err)
				break
			}

			msg := resp.Message
			history = append(history, msg)
			// what does the ToolCalls mean?
			if len(msg.ToolCalls) == 0 {
				fmt.Printf("\nLlama: %s\n", msg.Content)
				break
			}

			// Handle tool calls
			for _, tc := range msg.ToolCalls {
				fmt.Printf("[Calling Tool: %s with %s]\n", tc.Function.Name, tc.Function.Arguments)

				var args map[string]any
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					log.Printf("Failed to parse tool arguments: %v", err)
					continue
				}

				result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
					Params: mcp.CallToolParams{
						Name:      tc.Function.Name,
						Arguments: args,
					},
				})

				var resultText string
				if err != nil {
					resultText = fmt.Sprintf("Error: %v", err)
				} else {
					// We assume simple text result for this demo
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
				history = append(history, ChatMessage{
					Role:    "tool",
					Content: resultText,
				})
			}
			// Let Ollama process the tool results in the next iteration of the inner loop
		}
	}
}
