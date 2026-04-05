# ACP AG-UI Bridge

A Go-based bridge that connects ACP (Agent Client Protocol) compatible agents to the AG-UI protocol, exposing them via HTTP with a web interface.

## Overview

This project allows you to run any ACP-compatible agent (like Claude, OpenCode, Gemini, etc.) and interact with it through a web browser using the AG-UI protocol.

## Requirements

- Go 1.21+
- An ACP-compatible agent binary in your PATH

## Installation

```bash
cd examples/go-acp-agui-bridge
go build -o go-acp-agui-bridge .
```

## Usage

### Basic Usage

```bash
# Run with an ACP-compatible agent
./go-acp-agui-bridge --agent claude --port 3000

# Or with other agents
./go-acp-agui-bridge --agent opencode --port 8080
./go-acp-agui-bridge --agent gemini --port 4000
```

### Command Line Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--agent` | `-a` | (required) | Agent command to run |
| `--agent-args` | | `["--acp"]` | Agent arguments |
| `--port` | `-p` | `3000` | Server port |
| `--cwd` | | (current dir) | Working directory |
| `--verbose` | `-v` | `false` | Enable verbose logging |

### Accessing the Interface

Once running:
- **Web UI**: Open `http://localhost:3000` in your browser
- **AG-UI API**: Send POST requests to `http://localhost:3000/agent`

## AG-UI API

### Endpoint

```
POST /agent
Accept: text/event-stream
Content-Type: application/json
```

### Request Body

```json
{
  "threadId": "thread_123",
  "runId": "run_456",
  "messages": [
    { "role": "user", "content": "Hello!" }
  ]
}
```

### Response Events

The server streams AG-UI events:

- `RUN_STARTED` - Run initiated
- `TEXT_MESSAGE_START` - Assistant message begins
- `TEXT_MESSAGE_CONTENT` - Text delta from model
- `TEXT_MESSAGE_END` - Message complete
- `REASONING_MESSAGE_CONTENT` - Reasoning/thinking
- `TOOL_CALL_START` - Tool invocation
- `RUN_FINISHED` - Run complete
- `RUN_ERROR` - Error occurred

## Examples

### Using with Claude

```bash
./go-acp-agui-bridge --agent claude --agent-args "--acp"
```

### Using with OpenCode

```bash
./go-acp-agui-bridge --agent opencode --agent-args "acp"
```

### Custom Agent Arguments

```bash
./go-acp-agui-bridge --agent myagent --agent-args "--acp --model claude-sonnet"
```

## Architecture

```
┌─────────────┐     ACP (stdio)     ┌──────────────┐
│  Web UI     │◄───────────────────►│  Bridge      │◄──► ACP Agent
│             │     HTTP/SSE        │  (this)      │     (subprocess)
└─────────────┘                     └──────────────┘
```

1. Web browser connects to the bridge via HTTP
2. Bridge launches the ACP agent as a subprocess
3. Bridge translates between AG-UI events (HTTP/SSE) and ACP protocol (JSON-RPC/stdio)

## Files

- `main.go` - Main application code
- `static/index.html` - Web UI
- `static/index.js` - CopilotKit web component (bundled)

## See Also

- [ACP Protocol Documentation](../../_docs/acp.md)
- [AG-UI Protocol Documentation](../../_docs/agui.md)
- [go-acp-cli](../go-acp-cli) - ACP CLI client example
- [go-copilotkit](../go-copilotkit) - AG-UI/CopilotKit server example