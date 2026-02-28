package main

import (
	"context"
	"fmt"
	"strings"
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

	// ─────────────────────────────────────────────────────────────────────────
	// Reddit tools: generic, query-driven responses
	// ─────────────────────────────────────────────────────────────────────────

	// search_reddit: accepts any query and optional subreddit,
	// returns a structured Reddit-style summary for that specific topic.
	searchRedditTool := mcp.NewTool("search_reddit",
		mcp.WithDescription(
			"Search Reddit for community discussions, opinions, and insights on any topic. "+
				"Returns a curated summary of top threads and comments from relevant subreddits. "+
				"Use this when the user asks about experiences, opinions, recommendations, or community knowledge on any subject.",
		),
		mcp.WithString("query",
			mcp.Description("The topic or question to search for on Reddit (e.g. 'best way to learn Go', 'moving to Tokyo', 'budget travel Europe')"),
			mcp.Required(),
		),
		mcp.WithString("subreddit",
			mcp.Description("Optional: limit search to a specific subreddit (e.g. 'golang', 'travel', 'personalfinance'). Leave empty to search all of Reddit."),
		),
	)

	s.AddTool(searchRedditTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		query := args["query"].(string)
		subreddit, _ := args["subreddit"].(string)

		scope := "all of Reddit"
		scopeTag := "r/AskReddit + r/NoStupidQuestions + r/explainlikeimfive"
		if subreddit != "" {
			scope = fmt.Sprintf("r/%s", subreddit)
			scopeTag = fmt.Sprintf("r/%s", subreddit)
		}

		// Build a dynamic, query-aware summary
		result := fmt.Sprintf(
			"[Reddit Search] Scope: %s | Query: \"%s\"\n"+
				"Searched: %s\n\n"+
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
				"📌 Top Threads Found\n"+
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
				"🔺 Thread 1 — \"%s: What's the community consensus?\" (28.4k upvotes)\n"+
				"   ▸ Top comment [u/ExperiencedUser, 9.2k↑]:\n"+
				"     \"Having dealt with this myself, the most important thing is context.\n"+
				"     There's no one-size-fits-all answer, but the community generally agrees on a few core points.\"\n"+
				"   ▸ [u/DetailOriented, 6.7k↑]:\n"+
				"     \"Really depends on your specific situation. The factors that matter most are:\n"+
				"     (1) your background/experience, (2) your goals, and (3) available resources.\"\n"+
				"   ▸ [u/PracticalAdvice, 5.1k↑]:\n"+
				"     \"I've seen this asked many times. The short answer: start with the basics,\n"+
				"     build up gradually, and don't skip the fundamentals.\"\n\n"+
				"🔺 Thread 2 — \"Deep dive: everything you need to know about %s\" (14.1k upvotes)\n"+
				"   ▸ [u/SubjectMatterExpert, 8.3k↑]:\n"+
				"     \"After years of experience with this, I'd say the #1 mistake people make\n"+
				"     is overthinking it. Keep it simple and iterate.\"\n"+
				"   ▸ [u/ResearchNerd, 4.9k↑]:\n"+
				"     \"There's actually a lot of conflicting info out there. Here's what the\n"+
				"     evidence actually supports: consistency beats intensity almost every time.\"\n\n"+
				"🔺 Thread 3 — \"Beginner asking about %s — got great answers\" (7.8k upvotes)\n"+
				"   ▸ [u/HelpfulStranger, 3.4k↑]:\n"+
				"     \"Welcome to the rabbit hole! The community here is incredibly helpful.\n"+
				"     My advice: lurk for a week, read the wiki/FAQ, then ask specific questions.\"\n"+
				"   ▸ [u/VeteranMember, 2.1k↑]:\n"+
				"     \"Classic question. Check the sidebar resources — they cover exactly this.\"\n\n"+
				"📊 Community Consensus on \"%s\"\n"+
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
				"  • Most upvoted advice: Start simple, build on experience\n"+
				"  • Common pitfall flagged: Looking for a single 'right answer' — context matters\n"+
				"  • Recommended next step: Check the subreddit wiki and top posts of all time\n\n"+
				"🗂️ Best Reddit Forums for \"%s\"\n"+
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
				"  These are the subreddits where you'll get the highest quality answers:\n"+
				"  🥇 r/AskReddit          — 42M members | Broad opinions & personal experiences\n"+
				"  🥇 r/NoStupidQuestions  — 4.1M members | Beginner-friendly, judgment-free\n"+
				"  🥇 r/explainlikeimfive  — 22M members | Simple, clear explanations\n"+
				"  🥈 r/answers            — 850k members | Direct Q&A, fast responses\n"+
				"  🥈 r/OutOfTheLoop       — 3.2M members | Great for context on trending topics\n"+
				"  🥈 r/IAmA              — 21M members | Expert AMAs on specialized subjects\n"+
				"  💡 Tip: Use ask_reddit_community with any of the above for focused answers.\n\n"+
				"[Simulated | Retrieved: %s]",
			scope, query, scopeTag,
			query, query, query, query, query,
			time.Now().Format("2006-01-02 15:04 MST"),
		)

		return mcp.NewToolResultText(result), nil
	})

	// ask_reddit_community: post any question to a specific subreddit,
	// get back a realistic set of top community answers.
	askRedditTool := mcp.NewTool("ask_reddit_community",
		mcp.WithDescription(
			"Get community insights and answers from a specific subreddit on any question. "+
				"Simulates the top replies you'd receive from real Redditors in that community. "+
				"Use this when the user wants opinions, recommendations, or experiences from a focused community.",
		),
		mcp.WithString("subreddit",
			mcp.Description("The subreddit to ask (e.g. 'golang', 'AskReddit', 'personalfinance', 'travel', 'explainlikeimfive')"),
			mcp.Required(),
		),
		mcp.WithString("question",
			mcp.Description("The full question to post to the subreddit"),
			mcp.Required(),
		),
	)

	s.AddTool(askRedditTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		subreddit := args["subreddit"].(string)
		question := args["question"].(string)

		// Truncate question for display if long
		shortQ := question
		if len(shortQ) > 80 {
			shortQ = shortQ[:77] + "..."
		}

		result := fmt.Sprintf(
			"[r/%s] Posted: \"%s\"\n\n"+
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
				"🏆 Top Community Answers\n"+
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
				"💬 u/TopContributor_%s [18.3k↑]:\n"+
				"\"Great question for this sub. The honest answer is: it depends on your situation.\n"+
				"That said, the majority experience here points in one direction — see the thread\n"+
				"for more nuanced breakdowns.\"\n\n"+
				"💬 u/DetailedExplainer [11.7k↑]:\n"+
				"\"I'll give you the full breakdown:\n"+
				"  1. First, understand the underlying principles — don't skip this step.\n"+
				"  2. Look at what others in similar situations have done (this subreddit has great examples).\n"+
				"  3. Apply it to your specific context and iterate based on feedback.\n"+
				"Happy to go deeper if you share more details in the comments.\"\n\n"+
				"💬 u/VoiceOfReason [8.9k↑]:\n"+
				"\"Controversial take: most people overcomplicate this. The 80/20 rule applies here —\n"+
				"80%% of results come from 20%% of the effort. Focus on the fundamentals first.\"\n\n"+
				"💬 u/PersonalExperience [6.2k↑]:\n"+
				"\"Went through exactly this last year. What helped me most:\n"+
				"  • Reading the top posts of all time in r/%s\n"+
				"  • Asking follow-up questions in weekly threads\n"+
				"  • Finding a mentor or accountability partner\n"+
				"Good luck!\"\n\n"+
				"💬 u/SkepticalButHelpful [3.8k↑]:\n"+
				"\"Worth noting that a lot of advice here is anecdotal. Make sure to cross-reference\n"+
				"with primary sources when possible. That said, the collective wisdom in this thread\n"+
				"is actually pretty solid.\"\n\n"+
				"📊 Engagement Stats\n"+
				"  • 3,241 comments | 96%% upvoted | 🏅 Awarded ×5\n"+
				"  • Saved by 847 users | Cross-posted to r/BestOf\n\n"+
				"🔗 Related threads in r/%s:\n"+
				"  • \"Megathread: Resources and FAQ for common questions\"\n"+
				"  • \"Weekly discussion: share your experience\"\n"+
				"  • \"What I wish I knew when I started — community advice roundup\"\n\n"+
				"[Simulated | r/%s | Retrieved: %s]",
			subreddit, shortQ,
			strings.ToUpper(subreddit[:1])+subreddit[1:],
			subreddit, subreddit, subreddit,
			time.Now().Format("2006-01-02 15:04 MST"),
		)

		return mcp.NewToolResultText(result), nil
	})

	// list_subreddits: asks the LLM to use its own Reddit knowledge to recommend
	// real, topic-specific subreddits — no hardcoded keyword mappings.
	listSubredditsTool := mcp.NewTool("list_subreddits",
		mcp.WithDescription(
			"Find the most relevant Reddit communities (subreddits) for any topic. "+
				"Call this whenever the user asks about a topic and you want to recommend where on Reddit "+
				"they can get the best answers. Use your own knowledge of Reddit to suggest subreddits "+
				"that are real, active, and specifically focused on the topic — not just generic ones.",
		),
		mcp.WithString("topic",
			mcp.Description("The topic or subject to find subreddits for (e.g. 'studying abroad in Japan', 'learning Go', 'budget travel in Europe')"),
			mcp.Required(),
		),
	)

	s.AddTool(listSubredditsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		topic := args["topic"].(string)

		// Return a structured prompt so Ollama uses its own Reddit knowledge
		// to fill in the actual subreddit recommendations — no hardcoding needed.
		result := fmt.Sprintf(
			"[Subreddit Lookup] Topic: \"%s\"\n\n"+
				"Please use your knowledge of Reddit to recommend the most relevant subreddits for this topic.\n"+
				"For each subreddit, provide:\n"+
				"  - The subreddit name (r/...)\n"+
				"  - Approximate member count if you know it\n"+
				"  - A one-line description of what makes it useful for this specific topic\n\n"+
				"Prioritize specialized communities over generic ones (e.g. r/studyabroad is better than\n"+
				"r/AskReddit for study-abroad questions). List 4-6 subreddits ranked by relevance.\n\n"+
				"Also suggest which of those subreddits would be the single best place to post a question about:\n"+
				"\"%s\"",
			topic, topic,
		)

		return mcp.NewToolResultText(result), nil
	})

	return s
}
