# ACP Protocol Deep Dive: Functional Analysis

> Functional analysis of ACP protocol implementations based on source code from OpenCode, Cline, Codex-ACP, Claude-ACP, and Goose agents

---

## Overview

This document analyzes how prominent ACP-compatible agents implement the protocol, focusing on functional behaviors, message flows, and implementation patterns. We examine five agents to understand real-world patterns for authentication, session management, model selection, tool handling, and MCP integration.

---

## 1. Protocol Connection Model

### Functional Flow

1. **Client spawns agent** - Client launches agent as subprocess (stdio transport)
2. **Initialize exchange** - Client sends `initialize` request via stdin; agent responds with capabilities
3. **Authentication** - If required, client calls `authenticate` method
4. **Session creation** - Client creates session via `session/new` with cwd and optional MCP servers
5. **Interaction** - Proceeds via `session/prompt` requests and `session/update` notifications

**Reference:** ACP Protocol - Initialization, Session Setup sections

---

## 2. Initialization Behavior

### Initialize Request (Client → Agent)

Client sends:
- `protocolVersion` - version number
- `clientCapabilities` - what client supports (fs, terminal, auth)
- `clientInfo` - client name and version

### Initialize Response (Agent → Client)

Agent returns:
- `protocolVersion` - supported version
- `agentCapabilities` - what agent supports (loadSession, mcpCapabilities, promptCapabilities, sessionCapabilities)
- `authMethods` - available authentication methods
- `agentInfo` - agent name and version

### Capabilities Advertisement

All agents advertise MCP support via `mcpCapabilities`:
- `mcpCapabilities.http` - HTTP transport for MCP servers
- `mcpCapabilities.sse` - SSE transport (deprecated)

**Reference:** ACP Protocol - Initialization section

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - `initialize()` method (lines 535-578)
- Codex-ACP: `src/codex_agent.rs` - `initialize()` implementation (lines 218-254)
- Claude-ACP: `src/acp-agent.ts` - `initialize()` method

---

## 3. Authentication Flow

### Auth Methods Available

| Agent | Auth Method Type | Implementation |
|-------|-----------------|-----------------|
| OpenCode | Terminal | Custom command via terminal-auth capability |
| Claude-ACP | Terminal + Gateway | Subprocess login; custom endpoint routing |
| Codex-ACP | Agent + EnvVar | Browser OAuth; API key from env vars |
| Cline | Browser OAuth | Local server receives OAuth callback |

### Functional Flow

1. Agent returns `authMethods` in initialize response
2. If session creation fails with error code `-32001` (auth_required), client must call `authenticate`
3. After successful authentication, client retries session creation

**Error Code:** `-32001` (auth_required) signals authentication needed

**Reference:** ACP Protocol - Authentication section

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - handles LoadAPIKeyError, throws RequestError.authRequired()
- Codex-ACP: `src/codex_agent.rs` - `authenticate()` with CodexAuthMethod enum handling
- Claude-ACP: `src/acp-agent.ts` - gateway auth with custom base URL via `_meta`

---

## 4. Session Lifecycle

### Session Creation (session/new)

**Client sends:**
- `cwd` - working directory
- `mcpServers` - optional MCP servers to inject

**Agent behavior:**
1. Creates internal session/thread
2. Loads available models and modes
3. Registers any provided MCP servers

**Agent returns:**
- `sessionId` - session identifier
- `models` - available models list with current selection
- `modes` - available agent modes
- `configOptions` - session configuration options

### Session Loading (session/load)

**Client sends:**
- `sessionId` - existing session to resume
- `cwd` - current working directory
- `mcpServers` - MCP servers for this session

**Agent behavior:**
1. Reconstructs session from storage
2. Replays conversation history via `session_update` notifications
3. Returns current models/modes/config

**Key Pattern:** Session replay sends all previous messages as `user_message_chunk` or `agent_message_chunk` updates so client can rebuild UI

**Reference:** ACP Protocol - Session Setup section

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - `newSession()` (lines 585-617), `loadSession()` (lines 637-670)
- OpenCode: `packages/opencode/src/acp/session.ts` - ACPSessionManager class
- Codex-ACP: `src/codex_agent.rs` - Thread management with rollout history replay

---

## 5. Model Selection

### How Models Are Exposed

Agents return available models in session response:

```json
{
  "models": {
    "currentModelId": "anthropic/claude-sonnet-4-20250514",
    "availableModels": [
      { "id": "anthropic/claude-sonnet-4-20250514", "name": "Claude Sonnet 4" },
      { "id": "openai/gpt-4o", "name": "GPT-4o" }
    ]
  }
}
```

### Model System Variations

| Agent | Model System | Format |
|-------|-------------|--------|
| OpenCode | Provider/Model + Variants | `provider/model` with variant options |
| Claude-ACP | Alias Resolution | Human names like "opus" resolve to "claude-opus-4-6" |
| Codex-ACP | Presets + Effort | `{preset_id}/{effort}` like "o4-mini/low" |

### Runtime Model Change

Optional `setSessionModel` method allows changing model during session:
1. Client calls `setSessionModel` with `modelId`
2. Agent updates internal model selection
3. Returns confirmation with variant metadata

**Reference:** ACP Protocol - Session Setup, Schema sections

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - `setSessionModel()` (lines 1277-1297)
- Claude-ACP: `src/acp-agent.ts` - `resolveModelPreference()` for alias resolution
- Codex-ACP: `src/thread.rs` - ModelSelector with presets and reasoning effort

---

## 6. Prompt Interaction (session/prompt)

### Request Structure

**Client sends:**
- `sessionId` - active session
- `prompt` - array of ContentBlock (text, image, resource, resource_link)

### Response Flow

1. Agent processes prompt with selected model
2. Agent streams updates via `session_update` notifications
3. Agent returns final response with:
   - `stopReason` - why generation stopped (end_turn, max_tokens, cancelled, etc.)
   - `usage` - token usage information
   - `_meta` - additional metadata

### Content Block Types

| Type | Description | Usage |
|------|-------------|-------|
| `text` | Plain text | User messages, system prompts |
| `image` | Image data (base64 or URL) | Visual context |
| `resource_link` | File reference | Attach files by path |
| `resource` | Embedded content | Inline data (text or binary) |

**Key Feature:** `annotations.audience` can mark content for "assistant" only (synthetic) or mark content to ignore

**Reference:** ACP Protocol - Prompt Turn, Content sections

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - `prompt()` converts ContentBlock to internal format (lines 1308-1392)
- Claude-ACP: `src/acp-agent.ts` - main prompt loop with message processing

---

## 7. Session Update Notifications (Agent → Client)

| Update Type | Description | When Sent |
|-------------|-------------|-----------|
| `agent_message_chunk` | Text from model | Streaming text output |
| `agent_thought_chunk` | Reasoning/thinking | Model reasoning visibility |
| `user_message_chunk` | User message content | Message replay |
| `tool_call` | Tool call initiated | Model requests tool execution |
| `tool_call_update` | Tool progress/result | Tool running, completed, failed |
| `plan` | Task/plan updates | TODO or planning items |
| `usage_update` | Token usage | After message completion |
| `available_commands_update` | Slash commands | Session initialization |
| `current_mode_update` | Mode changed | When mode changes |

**Reference:** ACP Protocol - Prompt Turn, Tool Calls sections

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - `processMessage()` translates internal events (lines 825-1101)
- OpenCode: `packages/opencode/src/acp/agent.ts` - event subscription to SDK events (lines 167-182)
- Claude-ACP: `src/acp-agent.ts` - `query.next()` async iterator processing (lines 524-847)

---

## 8. Tool Call Handling

### Tool Call Flow

1. Model decides to call tool → sends `tool_call` update (status: pending)
2. Client may request permission via `session/request_permission` (optional)
3. Tool execution proceeds → `tool_call_update` (status: in_progress)
4. Tool completes → `tool_call_update` (status: completed or failed)

### Permission Handling

Agents implement permission systems:

| Agent | Permission Model |
|-------|------------------|
| OpenCode | Options: allow_once, allow_always, reject |
| Claude-ACP | Modes: auto, default, acceptEdits, plan, dontAsk, bypassPermissions |
| Codex-ACP | Guardian assessment with risk scores |

**Reference:** ACP Protocol - Tool Calls section

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - permission event handling (lines 186-265)
- Claude-ACP: `src/acp-agent.ts` - `canUseTool()` callback interface (lines 1068-1207)

---

## 9. MCP Server Integration

### MCP at Session Creation

Clients can inject MCP servers when creating session via `session/new` params:

```json
{
  "method": "session/new",
  "params": {
    "mcpServers": [
      { "type": "http", "name": "custom", "url": "http://localhost:8080", "headers": {} },
      { "type": "stdio", "name": "filesystem", "command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem", "/project"] }
    ]
  }
}
```

### Transport Support

| Transport | Parameter | Support |
|-----------|-----------|---------|
| Stdio | `{ command, args, env }` | All agents |
| HTTP | `{ url, headers }` | All agents |
| SSE | `{ url, headers }` | Limited/deprecated |

### Functional Behavior

1. Client provides MCP server config in session creation
2. Agent registers MCP servers with its MCP runtime
3. MCP server tools become available to the agent
4. Agent calls MCP tools → routed to client-provided server
5. Tool results returned to agent for continuation

**Reference:** ACP Protocol - MCP Integration section

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - converts McpServer to Config.Mcp (lines 1212-1250)
- Codex-ACP: `src/codex_agent.rs` - `build_session_config()` processes MCP (lines 120-213)

---

## 10. MCP Server Injection Pattern (Client as MCP Server)

### Overview

Clients can inject their own MCP servers at session creation, enabling bidirectional communication where client exposes functionality to agent.

### Functional Flow

1. **Client starts MCP server** - Runs HTTP server on localhost
2. **Session creation with MCP** - Passes server config in `session/new`
3. **Agent registers MCP** - Adds to internal MCP registry
4. **Tools available to agent** - Agent can call client-exposed tools
5. **Tool calls route to client** - Agent HTTP calls → client handler
6. **Client processes** - Can perform any action (UI updates, file ops, etc.)

### Use Cases

1. **UI Feedback** - Agent calls `show_notification`, `highlight_range`
2. **Client Capabilities** - `open_file`, `show_diff_preview`
3. **State Sync** - Agent reads/writes client state
4. **Protocol Bridge** - Client exposes AG-UI as MCP for agent interaction

### Architecture

```
┌─────────────┐   session/new+mcpServers   ┌─────────────┐
│   Client    │ ──────────────────────────▶│   Agent     │
│ (IDE/Editor)│                             │ (ACP Agent) │
│             │  ◀───────────────────────── │             │
│ [MCP Server]│   tool calls → responses    │ [MCP Client]│
│ localhost   │ ◀─────────────────────────── │             │
└─────────────┘   HTTP requests             └─────────────┘
```

**Reference:** ACP Protocol - Session Setup, MCP Integration sections

**Implementation References:**
- OpenCode: `packages/opencode/src/acp/agent.ts` - `newSession()`, `loadSession()` process mcpServers
- Codex-ACP: `src/codex_agent.rs` - `build_session_config()` converts MCP to internal format

---

## 11. Cross-Agent Comparison

| Aspect | OpenCode | Cline | Codex-ACP | Claude-ACP | Goose |
|--------|----------|-------|------------|------------|-------|
| **Language** | TypeScript | TypeScript | Rust | TypeScript | Rust |
| **Auth** | Terminal | Browser OAuth | ChatGPT + API keys | Terminal + Gateway | Agent auth |
| **Model System** | Provider/Variants | Permission modes | Presets+Effort | Alias resolution | Provider/Model |
| **Session** | ACPSessionManager | State map | Actor-based Thread | Pushable queue | ThreadManager |
| **MCP** | stdio+http | stdio+http | stdio+http | stdio+http | stdio+http |
| **Unique** | Reasoning chunks | Skills/Hooks | Guardian security | Enterprise settings | Extensions |

---

## 12. Key Implementation Patterns Summary

### Pattern 1: Initialize
- Returns protocol version, capabilities, auth methods, agent info

### Pattern 2: Session Creation
- Creates thread/session
- Loads models/modes/commands
- Returns sessionId with config

### Pattern 3: Prompt Loop
- Streams session updates for each message part
- Translates internal events to ACP session_update

### Pattern 4: Tool Handling
- Permission request → execute → stream result
- Maps internal tool states to ACP tool_call_update statuses

### Pattern 5: Error Translation
- Maps internal errors to ACP error codes
- Especially auth_required (code -32001)

---

## 13. References

### Primary Sources
- ACP Protocol: https://agentclientprotocol.com
- ACP Schema: https://agentclientprotocol.com/protocol/schema.md
- ACP Registry: https://github.com/agentclientprotocol/registry

### Agent Repositories
- OpenCode: https://github.com/anomalyco/opencode
- Cline: https://github.com/cline/cline
- Goose: https://github.com/block/goose
- Codex-ACP: https://github.com/zed-industries/codex-acp
- Claude-ACP: https://github.com/agentclientprotocol/claude-agent-acp

### Key Implementation References

**OpenCode:**
- `packages/opencode/src/acp/agent.ts` - Main ACP agent implementation
- `packages/opencode/src/acp/session.ts` - Session manager
- `packages/opencode/src/acp/types.ts` - Type definitions

**Codex-ACP:**
- `src/codex_agent.rs` - ACP agent trait implementation
- `src/thread.rs` - Thread/actor-based session management

**Claude-ACP:**
- `src/acp-agent.ts` - Main ACP agent wrapper
- `src/settings.ts` - Settings management
- `src/tools.ts` - Tool conversion logic