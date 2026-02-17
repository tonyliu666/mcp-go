package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewMCPServer() *server.MCPServer {
	// Create a new MCP server
	s := server.NewMCPServer("Llama MCP Demo Server", "1.0.0")

	// Add a simple calculation tool
	calculateTool := mcp.NewTool("calculate",
		mcp.WithDescription("Perform basic arithmetic operations"),
		mcp.WithString("operation",
			mcp.Description("The operation to perform (add, subtract, multiply, divide)"),
			mcp.Required(),
			mcp.Enum("add", "subtract", "multiply", "divide"),
		),
		mcp.WithNumber("a",
			mcp.Description("The first number"),
			mcp.Required(),
		),
		mcp.WithNumber("b",
			mcp.Description("The second number"),
			mcp.Required(),
		),
	)

	s.AddTool(calculateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		op := args["operation"].(string)
		a := args["a"].(float64)
		b := args["b"].(float64)

		var result float64
		switch op {
		case "add":
			result = a + b
		case "subtract":
			result = a - b
		case "multiply":
			result = a * b
		case "divide":
			if b == 0 {
				return mcp.NewToolResultError("Division by zero"), nil
			}
			result = a / b
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown operation: %s", op)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%f", result)), nil
	})

	// Add a get_time tool
	timeTool := mcp.NewTool("get_time",
		mcp.WithDescription("Get the current time"),
	)

	s.AddTool(timeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(time.Now().Format(time.RFC3339)), nil
	})

	return s
}
