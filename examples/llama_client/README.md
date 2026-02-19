### Architecture:

```text
┌─────────────────────────────────────────────────────────┐
│                      Ollama Server                      │
└─────────────────────────────────────────────────────────┘
          │                 │                │
          ▼                 ▼                ▼
  ┌───────────────┐ ┌───────────────┐ ┌────────────────┐
  │  MCP Clients  │ │  HTTP Server  │ │  Ollama Client │
  └───────┬───────┘ │ (MCP Server)  │ └────────┬───────┘
          │         └───────┬───────┘          │
          │ connects        │                  │ Request (1)
          └───────►         │                  ▼
          ◄─────────────────┘          ┌────────────────┐
             Get MCP Tools             │  Ollama Server │
                                       └────────┬───────┘
                                                │
                                                ▼
                                    Ask which tool should cope
                                       with the question?
                                                │
                                                ▼
                                       ┌────────────────┐
                                       │  Request (2):  │
                                       │ Execution Order│
                                       └────────────────┘
```

### How it works:

1. **MCP Server**: Runs locally on port 3000 and exposes MCP tools.
2. **Ollama Client**: Connects to the MCP Server to get tools.
3. **Ollama Server**: Receives tools from MCP Server and uses them to answer questions.

