### Architecture:

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