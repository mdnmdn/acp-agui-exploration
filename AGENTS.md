# ACP AGUI Bridge Project

> Research and implementation exploration for bridging Agent Client Protocol (ACP) and Agent User Interaction Protocol (AG-UI)

---

## Project Overview

This project explores the integration between two AI agent protocols:
- **ACP (Agent Client Protocol)** - Protocol for IDE/editor to agent communication (JSON-RPC over stdio)
- **AG-UI (Agent User Interaction Protocol)** - Protocol for frontend to agent communication (event-driven over HTTP/SSE)

The goal is to understand how these protocols work, their differences, and how to build bridges between them.

---

## Project Structure

```
acp-agui-bridge/
├── _docs/           # Protocol documentation and analysis
├── _examples/       # Implementation examples in various stacks
├── agents/          # Cloned agent repositories for analysis
├── wip/            # Work-in-progress code explorations
└── AGENTS.md       # This file
```

---

## _docs/ - Protocol Documentation

### Core Protocol Docs

| File | Description |
|------|-------------|
| `acp.md` | Agent Client Protocol full reference - transport, messages, authentication, error handling |
| `agui.md` | Agent User Interaction Protocol full reference - events, messages, tools, state |
| `acp-registry.md` | ACP Registry analysis - agents, distribution, authentication requirements |

### Deep Dive & Analysis

| File | Description |
|------|-------------|
| `acp-deepdive.md` | Functional analysis of ACP implementations from 5 agents (OpenCode, Cline, Codex-ACP, Claude-ACP, Goose) |
| `acp-to-agui-bridge.md` | Specification for ACP → AG-UI bridge |
| `agui-to-acp-bridge.md` | Specification for AG-UI → ACP bridge |

---

## _examples/ - Implementation Examples

Working examples in various tech stacks:

| Example | Stack | Description |
|---------|-------|-------------|
| [go-copilotkit](./examples/go-copilotkit/) | Go | AG-UI/CopilotKit server with Gemini model - implements SSE endpoint for browser clients |
| [go-acp-cli](./examples/go-acp-cli/) | Go | ACP CLI client with TUI - connects to ACP agents via stdio, includes mock agent |
| [go-acp-agui-bridge](./examples/go-acp-agui-bridge/) | Go | ACP to AG-UI bridge - runs ACP agents and exposes AG-UI protocol via HTTP with web UI |
| [copilotkit-webcomponent](./examples/copilotkit-webcomponent/) | TypeScript/React | Web component for embedding CopilotKit chat UI in any web page |

---

## agents/ - Agent Analysis

Cloned repositories of ACP-compatible agents for code analysis:

| Agent | Source | Type |
|-------|--------|------|
| OpenCode | https://github.com/anomalyco/opencode | Native ACP (TypeScript) |
| Cline | https://github.com/cline/cline | Native ACP (TypeScript) |
| Goose | https://github.com/block/goose | Native ACP (Rust) |
| Codex-ACP | https://github.com/zed-industries/codex-acp | Adapter (wraps OpenAI Codex) |
| Claude-ACP | https://github.com/agentclientprotocol/claude-agent-acp | Adapter (wraps Claude Agent SDK) |

These are used to analyze:
- How agents implement ACP
- Authentication patterns
- Model selection
- Tool handling
- MCP integration
- Session management

---

## wip/ - Work in Progress

Exploratory code and experiments:
- Bridge prototypes
- Protocol mapping experiments
- Test implementations

---

## Key Findings

### ACP Protocol
- Uses JSON-RPC 2.0 over stdio
- Session-based with session/new, session/prompt, session/update
- MCP servers can be injected at session creation
- Authentication via authMethods in initialize
- Tool permissions via session/request_permission

### AG-UI Protocol  
- Event-driven streaming over HTTP/SSE
- 16+ event types (RUN_STARTED, TEXT_MESSAGE_CONTENT, TOOL_CALL, etc.)
- State sync via STATE_SNAPSHOT/STATE_DELTA
- Frontend-defined tools

### Bridge Architecture
- ACP → AGUI: JSON-RPC requests → event stream
- AGUI → ACP: events → subprocess + JSON-RPC
- MCP injection enables bidirectional communication
- Client can act as MCP server to expose capabilities

---

## References

- [ACP Protocol](https://agentclientprotocol.com)
- [AG-UI Protocol](https://docs.ag-ui.com)
- [ACP Registry](https://cdn.agentclientprotocol.com/registry/v1/latest/registry.json)
- [MCP Protocol](https://modelcontextprotocol.io)