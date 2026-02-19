### Architecture:

graph TD
    OS1[Ollama Server] --> MC[MCP Clients]
    OS1 --> HS[HTTP Server / MCP Server]
    OS1 --> OC[Ollama Client]

    OC -- "Request (1)" --> OS2[Ollama Server]
    OS2 -- "Ask which tool?" --> RO["Request (2): Execution Order"]

    MC -- "Get MCP Tools" --- HS