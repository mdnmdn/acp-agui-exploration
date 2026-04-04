# ACP AGUI Bridge Exploration

Research and implementation exploration for bridging Agent Client Protocol (ACP) and Agent User Interaction Protocol (AG-UI).

## Project Overview

This project explores the integration between two AI agent protocols:
- **ACP (Agent Client Protocol)** - Protocol for IDE/editor to agent communication (JSON-RPC over stdio)
- **AG-UI (Agent User Interaction Protocol)** - Protocol for frontend to agent communication (event-driven over HTTP/SSE)

The goal is to understand how these protocols work, their differences, and how to build bridges between them.

## Project Structure

```
acp-agui-bridge/
├── _docs/           # Protocol documentation and analysis
├── _examples/       # Implementation examples in various stacks
├── agents/          # Cloned agent repositories for analysis
├── wip/            # Work-in-progress code explorations
└── README.md       # This file
```

## Key Documents

- [_docs/agui.md](_docs/agui.md) - Agent User Interaction Protocol reference
- [_docs/acp.md](_docs/acp.md) - Agent Client Protocol reference
- [_docs/agui-copilotkit-protocols.md](_docs/agui-copilotkit-protocols.md) - AG-UI and CopilotKit protocol analysis
- [examples/go-acp-cli/](examples/go-acp-cli/) - Go-based ACP client example

## Getting Started

### Go ACP CLI Example

The `examples/go-acp-cli` directory contains a working ACP client implementation:

```bash
cd examples/go-acp-cli
go build .
./go-acp-cli          # Launch TUI
./go-acp-cli -a mock "hello world"  # Use mock agent in CLI mode
```

## License

This project is open source and available under the MIT License.