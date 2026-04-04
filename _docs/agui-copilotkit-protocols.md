# AG-UI and CopilotKit Protocols

## Overview

This document explores the integration between Agent User Interaction Protocol (AG-UI) and CopilotKit, focusing on how CopilotKit implements the AG-UI protocol to enable communication between frontend applications and AI agents.

## CopilotKit Architecture

CopilotKit follows a layered architecture:

1. **Frontend Layer**: React components that provide UI for agent interactions
2. **Runtime Layer**: Core CopilotKit runtime that handles agent execution and protocol translation
3. **Agent Layer**: Individual AI agents that process requests and generate responses

### Key Components

- **CopilotRuntime**: Main runtime class that orchestrates agent execution
- **Adapter System**: Connects to various LLM providers (OpenAI, Gemini, etc.)
- **BuiltInAgent**: Base class for creating custom agents
- **HTTP Endpoint Handler**: Exposes AG-UI protocol over HTTP/SSE

## AG-UI Implementation in CopilotKit

CopilotKit implements AG-UI primarily through its runtime endpoints, specifically the `copilotRuntimeNodeHttpEndpoint` function.

### Request Handling Flow

1. **HTTP POST Request**: Frontend sends agent requests to `/copilotkit` endpoint
2. **Protocol Negotiation**: Based on `Accept` header, determines transport (SSE, binary, etc.)
3. **Agent Execution**: Runtime invokes appropriate agent with provided input
4. **Event Streaming**: Agent output is converted to AG-UI events and streamed back

### Supported Transports

CopilotKit supports the standard AG-UI transports:
- **Server-Sent Events (SSE)**: Default for browser compatibility
- **Binary Protocol**: For high-volume scenarios
- **WebSocket**: For bidirectional persistent connections
- **Webhook**: For server-to-server communication

## CopilotKit Server Implementation Analysis

Based on `wip/reference-copilotkit-runtime/server.ts`:

### Core Implementation Details

```typescript
const runtime = new CopilotRuntime({
    agents: {
        default: new BuiltInAgent({
            model: "gemini/gemini-flash-lite-latest",
            tools: [roastTool],
        }),
        mathGuy: new BuiltInAgent({
            model: "gemini/gemini-flash-lite-latest",
            tools: [randomNumberTool],
        }),
    },
    remoteEndpoints: [
        { url: "http://localhost:8000/copilotkit" },
    ],
    observability: {
        enabled: true,
        progressive: false,
        hooks: {
            handleRequest: async (data) => {
                console.log("handleRequest", data);
            },
            handleError: async (data) => {
                console.error("handleError", data);
            },
            handleResponse: async (data) => {
                console.log("handleResponse", data);
            }
        }
    }
});

const handler = copilotRuntimeNodeHttpEndpoint({
    endpoint: '/copilotkit',
    runtime,
});

const server = createServer(logMiddleware((req, res) => {
    handler(req, res);
}));
```

### Key Features Observed

1. **Multi-agent Support**: Runtime can manage multiple named agents
2. **Tool Integration**: Agents can be configured with specific tools
3. **Remote Endpoints**: Ability to proxy to other CopilotKit instances
4. **Observability Hooks**: Request/response/error logging capabilities
5. **Middleware Support**: Custom middleware for logging/authentication

## Go Implementation Analysis

Based on `wip/go-copilotkit/protocol.go`:

### Protocol Structure

The Go implementation demonstrates a clean separation of concerns:

1. **Protocol Struct**: Main handler that coordinates agent and storage
2. **Agent Interface**: Defines required methods for agent implementations
3. **SSE Writer**: Handles Server-Sent Events formatting and flushing
4. **Request Routing**: Dispatches based on method (info, agent/connect, agent/run)

### Event Streaming Implementation

The Go code shows proper AG-UI SSE implementation:

```go
// Proper SSE framing with data: prefix and double newline
_, err = s.w.Write([]byte("data: " + string(jsonData) + "\n\n"))

// Correct headers for SSE
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")
w.Header().Set("Connection", "keep-alive")

// Standard AG-UI event sequence
stream.Write(map[string]interface{}{
    "type":     "RUN_STARTED",
    "threadId": body.ThreadID,
    "runId":    body.RunID,
})
```

### Agent Lifecycle Management

1. **Connect**: Handles initial agent connection (minimal in this implementation)
2. **Run**: Main execution logic that:
   - Loads thread history
   - Saves new user messages
   - Streams RUN_STARTED event
   - Streams TEXT_MESSAGE_START
   - Executes agent logic via collector
   - Streams TEXT_MESSAGE_END
   - Streams RUN_FINISHED

## Protocol Comparison: AG-UI vs Standard CopilotKit

### Similarities

1. **Event-Based Communication**: Both use typed events for communication
2. **Streaming Responses**: Support for real-time response streaming
3. **Tool Execution**: Both support tool calling mechanisms
4. **State Management**: Capabilities for maintaining conversation state

### Differences

1. **Transport Specificity**: AG-UI is transport-agnostic; standard CopilotKit may have tighter HTTP coupling
2. **Event Standardization**: AG-UI defines specific event types; CopilotKit may use proprietary formats
3. **Client SDKs**: AG-UI has official client SDKs; standard CopilotKit may have different client libraries
4. **Middleware Approach**: AG-UI middleware is more standardized across implementations

## Implementation Recommendations

### For ACP ↔ AG-UI Bridging

1. **Use CopilotKit Runtime**: Leverage existing AG-UI implementation rather than building from scratch
2. **Adapter Pattern**: Create ACP agent that translates to/from AG-UI events
3. **Session Mapping**: Map ACP sessions to AG-UI thread/run IDs
4. **Tool Translation**: Convert between ACP tool calls and AG-UI tool events
5. **State Synchronization**: Implement state delta translation between protocols

### Key Integration Points

1. **ACP session/new → AG-UI thread initialization**
2. **ACP session/prompt → AG-UI RUN_AGENT request**
3. **ACP tool calls ↔ AG-UI TOOL_CALL events**
4. **ACP session updates ↔ AG-UI STATE_DELTA events**
5. **ACP notifications ↔ AG-UI RAW/CUSTOM events**

## Conclusion

CopilotKit provides a robust implementation of the AG-UI protocol that can be leveraged for bridging with ACP. The runtime architecture is modular and extensible, making it suitable for protocol translation tasks. The Go reference implementation demonstrates how to properly implement AG-UI SSE streaming, while the TypeScript example shows how to configure and extend the runtime for specific use cases.

For ACP ↔ AG-UI bridging, the focus should be on:
1. Creating adapters that translate between protocol-specific messages
2. Maintaining state consistency between the two systems
3. Properly handling error cases and edge conditions
4. Leveraging existing CopilotKit runtime capabilities rather than reimplementing core functionality