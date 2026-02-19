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

	// Add a tool to search Reddit (useful for gathering information to answer questions)
	searchRedditTool := mcp.NewTool("search_reddit",
		mcp.WithDescription("Search Reddit for threads and comments (Simulated)"),
		mcp.WithString("query",
			mcp.Description("The search query"),
			mcp.Required(),
		),
		mcp.WithString("subreddit",
			mcp.Description("Optional subreddit to search within"),
		),
	)

	s.AddTool(searchRedditTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		query := args["query"].(string)
		subreddit, _ := args["subreddit"].(string)

		subText := "all of Reddit"
		if subreddit != "" {
			subText = fmt.Sprintf("r/%s", subreddit)
		}

		// Simulated search results with specific user-requested facts
		result := fmt.Sprintf("[Reddit Search Simulation] Searching %s for: '%s'\n\n"+
			"Top Results:\n"+
			"1. [r/AskReddit] 'What is your most mildly interesting secret?' (15k upvotes)\n"+
			"   - 'I unknowingly created numerous crop circles in my village, which led to police patrols and UFO hunter comments. I had no idea that it was illegal and didn't really interact with the rest of the village.'\n"+
			"2. [r/Parenting] 'What weird skills did you have as a kid?' (2.4k upvotes)\n"+
			"   - 'When I was in 2nd and 1st grade I could sing the entire alphabet with dinosaurs. Different dinosaur for each letter!'\n"+
			"3. [r/%s] 'Discussion on %s' (800 upvotes) - 'The consensus seems to be that most users recommend checking the official docs.'\n\n"+
			"Summary: Reddit is full of both useful advice and bizarre personal anecdotes.",
			subText, query, subreddit, query)
		return mcp.NewToolResultText(result), nil
	})

	// Add a tool to ask a question to a subreddit
	askRedditTool := mcp.NewTool("ask_reddit_question",
		mcp.WithDescription("Post a question to a specific subreddit (Simulated)"),
		mcp.WithString("subreddit",
			mcp.Description("The name of the subreddit (e.g., 'golang', 'programming')"),
			mcp.Required(),
		),
		mcp.WithString("question",
			mcp.Description("The question to ask"),
			mcp.Required(),
		),
	)

	s.AddTool(askRedditTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		subreddit := args["subreddit"].(string)
		question := args["question"].(string)

		// Enhancing the simulation with "recommended content" from other Redditors
		result := fmt.Sprintf("[Reddit Simulation] Posted to r/%s: %s\n\n"+
			"Status: Pending Approval/Moderation.\n"+
			"Initial Automated Feedback: Your post fits the community guidelines for r/%s.\n\n"+
			"Trending Fun Facts from Redditors:\n"+
			"- The 'Crop Circle Creator': A Redditor once created crop circles in their village by accident, sparking UFO rumors!\n"+
			"- The 'Dinosaur Alphabet': One user could sing the entire alphabet using only dinosaur names.\n\n"+
			"Expected Interaction: Similar posts in r/%s usually get 10-20 comments in the first hour.",
			subreddit, question, subreddit, subreddit)
		return mcp.NewToolResultText(result), nil
	})

	return s
}
