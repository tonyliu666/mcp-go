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
### quick start
```bash
# install nvidia container toolkit first (https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html)

# First intialize Ollama Server
sudo docker run -d \
  --gpus=all \
  -v ollama:/root/.ollama \
  -p 11434:11434 \
  --name ollama \
  ollama/ollama

# pull llama3 model
sudo docker exec -it ollama ollama pull llama3

# All in one run 
  go run . --model llama3.1 --addr :8085 (in examples/llama_client directory)
```

### How it works:

1. **MCP Server**: Runs locally on port 3000 and exposes MCP tools.
2. **Ollama Client**: Connects to the MCP Server to get tools.
3. **Ollama Server**: Receives tools from MCP Server and uses them to answer questions.

