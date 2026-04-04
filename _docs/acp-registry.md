# ACP Registry Analysis

> Comprehensive analysis of the Agent Client Protocol Registry, agents, and implementation patterns

## Overview

The **ACP Registry** is a curated directory of agents implementing the Agent Client Protocol (ACP). It provides a standardized way to discover, distribute, and integrate AI coding agents into ACP-compatible clients (like Zed, Cursor, JetBrains IDEs).

**Registry URL:** `https://cdn.agentclientprotocol.com/registry/v1/latest/registry.json`

**Key Requirement:** All agents must support user authentication to be included.

---

## Registry Structure

### Schema

```json
{
  "version": "1.0.0",
  "agents": [...],
  "extensions": []
}
```

### Agent Entry Schema

```json
{
  "id": "agent-id",           // Unique identifier (lowercase, hyphens)
  "name": "Display Name",     // Human-readable name
  "version": "1.0.0",         // Semantic version
  "description": "...",       // Brief description
  "repository": "https://github.com/...",
  "website": "https://...",
  "authors": ["Author Name"],
  "license": "MIT",
  "icon": "https://.../agent-id.svg",
  "distribution": {
    "binary": { ... },        // Platform-specific executables
    "npx": { ... },           // npm packages
    "uvx": { ... }            // PyPI packages via uv
  }
}
```

---

## Distribution Methods

| Type | Description | Command | Example |
|------|-------------|---------|---------|
| **binary** | Platform-specific executables | Download, extract, run | `./agent --acp` |
| **npx** | npm packages | `npx <package> [args]` | `npx @scope/agent --acp` |
| **uvx** | PyPI packages via uv | `uvx <package> [args]` | `uvx agent-package` |

### Supported Platforms (binary)

- `darwin-aarch64` - macOS Apple Silicon
- `darwin-x86_64` - macOS Intel
- `linux-aarch64` - Linux ARM64
- `linux-x86_64` - Linux x86_64
- `windows-aarch64` - Windows ARM64
- `windows-x86_64` - Windows x86_64

### Supported Archive Formats

**Supported:** `.zip`, `.tar.gz`, `.tgz`, `.tar.bz2`, `.tbz2`, raw binaries

**Not Supported:** `.dmg`, `.pkg`, `.deb`, `.rpm`, `.msi`, `.appimage` (installer formats)

---

## Registry Agents (Current)

> **Classification:** Agents are classified as either **Native ACP** (built-in ACP support) or **Adapter/Wrapper** (translate another protocol to ACP)

| Agent | Version | Distribution | License | Type | Wraps |
|-------|---------|--------------|---------|------|-------|
| **Amp** | 0.7.0 | binary (all) | Apache-2.0 | Native ACP | - |
| **Auggie CLI** | 0.22.0 | npx | proprietary | Native ACP | - |
| **Autohand Code** | 0.2.1 | npx | Apache-2.0 | Native ACP | - |
| **Claude Agent** | 0.25.0 | npx | proprietary | Adapter | Claude Agent SDK |
| **Cline** | 2.13.0 | npx | Apache-2.0 | Native ACP | - |
| **Codebuddy Code** | 2.77.0 | npx | proprietary | Native ACP | - |
| **Codex CLI** | 0.11.1 | binary + npx | Apache-2.0 | Adapter | OpenAI Codex |
| **Corust Agent** | 0.4.1 | binary (4) | GPL-3.0 | Native ACP | - |
| **crow-cli** | 0.1.14 | uvx | Apache-2.0 | Native ACP | - |
| **Cursor** | 2026.03.30 | binary (all) | proprietary | Native ACP | - |
| **DeepAgents** | 0.1.7 | npx | MIT | Native ACP | - |
| **DimCode** | 0.0.20 | npx | proprietary | Native ACP | - |
| **Factory Droid** | 0.93.0 | npx | proprietary | Native ACP | - |
| **fast-agent** | 0.6.10 | uvx | Apache-2.0 | Native ACP | - |
| **Gemini CLI** | 0.36.0 | npx | Apache-2.0 | Native ACP | - |
| **GitHub Copilot** | 1.0.17 | npx | proprietary | Native ACP | - |
| **goose** | 1.29.1 | binary (5) | Apache-2.0 | Native ACP | - |
| **Junie** | 888.212.0 | binary (5) | proprietary | Native ACP | - |
| **Kilo** | 7.1.20 | binary + npx | MIT | Native ACP | - |
| **Kimi CLI** | 1.30.0 | binary (4) | MIT | Native ACP | - |
| **Minion Code** | 0.1.44 | uvx | AGPL-3.0 | Native ACP | - |
| **Mistral Vibe** | 2.7.3 | binary (6) | Apache-2.0 | Native ACP | - |
| **Nova** | 1.0.93 | npx | proprietary | Native ACP | - |
| **OpenCode** | 1.3.13 | binary (5) | MIT | Native ACP | - |
| **pi ACP** | 0.0.24 | npx | MIT | Adapter | Pi agent |
| **Qoder CLI** | 0.1.38 | npx | proprietary | Native ACP | - |
| **Qwen Code** | 0.14.0 | npx | Apache-2.0 | Native ACP | - |
| **Stakpak** | 0.3.71 | binary (5) | Apache-2.0 | Native ACP | - |

**Total: 29 agents**
- **Native ACP:** 25 agents
- **Adapter/Wrapper:** 3 agents (claude-acp, codex-acp, pi-acp)

---

## Agent Type Definitions

### Native ACP Agents

Agents that implement ACP protocol directly from the ground up. They typically:
- Have a built-in `--acp` or `acp` flag to enable ACP mode
- Implement the full ACP Agent interface directly
- Handle JSON-RPC communication natively

### Adapter/Wrapper Agents

Agents that translate another protocol (or agent SDK) to ACP. They:
- Wrap an existing agent/CLI and convert between protocols
- Use the underlying agent's SDK while exposing ACP interface
- Enable using non-ACP-native agents via ACP

---

## Authentication Requirements

To be included in the registry, agents **must support authentication**. The registry verifies this via CI.

### Supported Auth Methods

1. **Agent Auth** - Agent handles OAuth flow itself
2. **Terminal Auth** - Interactive TUI-based authentication

### Agent Auth (Default)

```json
{
  "id": "agent",
  "name": "Agent",
  "description": "Authenticate through agent",
  "type": "agent"
}
```

**Flow:**
1. Client triggers agent's auth flow
2. Agent starts local HTTP server for OAuth callback
3. Agent opens user's browser with OAuth URL
4. User authenticates in browser
5. Provider redirects back with auth code
6. Agent exchanges code for tokens
7. Agent stores credentials securely

### Terminal Auth

```json
{
  "id": "terminal-auth",
  "name": "Run in terminal",
  "type": "terminal",
  "args": ["--setup"],
  "env": { "VAR1": "value1" }
}
```

**Flow:**
1. Client launches agent with setup args
2. Agent presents interactive terminal UI
3. User completes auth in terminal
4. Agent ready for standard ACP communication

---

## Common Implementation Patterns

### 1. Native ACP Mode

Many agents have built-in ACP support with a native flag:

```json
{
  "distribution": {
    "npx": {
      "package": "cline@2.13.0",
      "args": ["--acp"]
    }
  }
}
```

The `--acp` flag tells the agent to speak ACP instead of its native protocol. These are **Native ACP** agents.

### 2. Adapter Pattern (Wrapper)

Adapters wrap an external agent/SDK and translate between protocols:

```json
{
  "distribution": {
    "npx": {
      "package": "@agentclientprotocol/claude-agent-acp",
      "args": ["--acp"]
    }
  }
}
```

This pattern is used by:
- `claude-acp` - Wraps Claude Agent SDK
- `codex-acp` - Wraps OpenAI Codex CLI
- `pi-acp` - Wraps Pi agent

### 3. Binary + Package Hybrid

Some agents provide both binary and npm distributions:

```json
{
  "distribution": {
    "binary": {
      "darwin-aarch64": { "archive": "...", "cmd": "./agent", "args": ["acp"] },
      "linux-x86_64": { "archive": "...", "cmd": "./agent", "args": ["acp"] }
    },
    "npx": {
      "package": "@scope/agent@1.0.0",
      "args": ["acp"]
    }
  }
}
```

### 3. Custom Entry Points

Different distribution types may use different commands:

```json
{
  "distribution": {
    "npx": {
      "package": "@compass-ai/nova@1.0.93",
      "args": ["acp"]
    }
  }
}
```

```json
{
  "distribution": {
    "binary": {
      "darwin-aarch64": {
        "archive": "https://downloads.cursor.com/.../agent-cli-package.tar.gz",
        "cmd": "./dist-package/cursor-agent",
        "args": ["acp"]
      }
    }
  }
}
```

---

## Version Management

### Automatic Updates

Versions are **automatically updated hourly** via cron job:
- Checks npm for latest published version
- Checks PyPI for latest published version
- Checks GitHub Releases for latest tag and assets
- Commits updates directly to `main`

### Manual Updates

For non-GitHub releases or custom changes:
1. Fork the registry
2. Update `agent.json`
3. Submit PR
4. CI validates and merges
5. New release is published

---

## Validation Requirements

### Schema Validation
- Valid JSON against `agent.schema.json`

### ID Validation
- Lowercase letters, digits, hyphens only
- Must start with a letter
- Must match directory name
- Must be unique

### Version Validation
- Semantic version format (`x.y.z`)
- All parts must be numeric

### Distribution Validation
- At least one distribution method required
- Binary: requires `archive` and `cmd` per platform
- Package: requires `package` field

### Icon Validation
- **Exactly 16x16 SVG** (via width/height or viewBox)
- **Monochrome** using `currentColor`
- No hardcoded colors (#FF0000, red, rgb(), etc.)

### Authentication Validation
- Agent must return `authMethods` in `initialize` response
- At least one method must be `type: "agent"` or `type: "terminal"`

---

## Client Implementation Example

### Fetching and Using the Registry

```typescript
// Fetch registry
const response = await fetch(
  'https://cdn.agentclientprotocol.com/registry/v1/latest/registry.json'
);
const registry = await response.json();

// Filter by platform and distribution
const agents = registry.agents.filter(agent => {
  // Check if binary available for your platform
  if (agent.distribution.binary) {
    return agent.distribution.binary['darwin-aarch64'] !== undefined;
  }
  // Or check npx/uvx availability
  return agent.distribution.npx !== undefined;
});

// Get installation command
function getInstallCommand(agent, platform = 'darwin-aarch64') {
  if (agent.distribution.binary?.[platform]) {
    const { archive, cmd, args = [] } = agent.distribution.binary[platform];
    return { type: 'binary', archive, cmd, args };
  }
  if (agent.distribution.npx) {
    return { type: 'npx', ...agent.distribution.npx };
  }
  if (agent.distribution.uvx) {
    return { type: 'uvx', ...agent.distribution.uvx };
  }
}
```

---

## Adding an Agent to the Registry

1. **Fork** the registry repository
2. **Create directory**: `mkdir <agent-id>/`
3. **Create `agent.json`** with required fields
4. **Add icon**: `<agent-id>/icon.svg` (16x16, monochrome)
5. **Submit PR** - CI validates automatically

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier |
| `name` | string | Display name |
| `version` | string | Semantic version |
| `description` | string | Brief description |
| `distribution` | object | At least one method |

### Optional Fields

| Field | Type |
|-------|------|
| `repository` | string |
| `website` | string |
| `authors` | array |
| `license` | string |

---

## Key References

- **Registry Repository:** https://github.com/agentclientprotocol/registry
- **Registry CDN:** https://cdn.agentclientprotocol.com/registry/v1/latest/registry.json
- **Format Spec:** https://github.com/agentclientprotocol/registry/blob/main/FORMAT.md
- **Auth Spec:** https://github.com/agentclientprotocol/registry/blob/main/AUTHENTICATION.md
- **Contributing:** https://github.com/agentclientprotocol/registry/blob/main/CONTRIBUTING.md
- **ACP Protocol:** https://agentclientprotocol.com
- **ACP Auth RFD:** https://agentclientprotocol.com/rfds/auth-methods

---

# Using the Registry in a Custom Client

This section describes how to integrate the ACP Registry into your own client application (IDE, editor, or tool).

## Overview

A typical client workflow:
1. Fetch the registry JSON from CDN
2. Filter agents by your platform and preferences
3. Present available agents to the user
4. Download/install the selected agent
5. Launch the agent as an ACP subprocess
6. Initialize the ACP connection

---

## Step 1: Fetch the Registry

### JavaScript/TypeScript

```typescript
interface Registry {
  version: string;
  agents: Agent[];
  extensions: unknown[];
}

interface Agent {
  id: string;
  name: string;
  version: string;
  description: string;
  repository?: string;
  website?: string;
  authors: string[];
  license: string;
  icon?: string;
  distribution: Distribution;
}

interface Distribution {
  binary?: Record<string, BinaryTarget>;
  npx?: NpxTarget;
  uvx?: UvxTarget;
}

interface BinaryTarget {
  archive: string;
  cmd: string;
  args?: string[];
  env?: Record<string, string>;
}

async function fetchRegistry(): Promise<Registry> {
  const response = await fetch(
    'https://cdn.agentclientprotocol.com/registry/v1/latest/registry.json'
  );
  
  if (!response.ok) {
    throw new Error(`Failed to fetch registry: ${response.status}`);
  }
  
  return response.json();
}
```

### Python

```python
import requests
from dataclasses import dataclass
from typing import List, Optional, Dict, Any

@dataclass
class Agent:
    id: str
    name: str
    version: str
    description: str
    authors: List[str]
    license: str
    repository: Optional[str] = None
    website: Optional[str] = None
    icon: Optional[str] = None
    distribution: Dict[str, Any] = None

def fetch_registry() -> Dict[str, Any]:
    url = "https://cdn.agentclientprotocol.com/registry/v1/latest/registry.json"
    response = requests.get(url, timeout=30)
    response.raise_for_status()
    return response.json()
```

---

## Step 2: Filter Agents by Platform

### Detect Current Platform

```typescript
function getCurrentPlatform(): string {
  const os = process.platform; // 'darwin', 'linux', 'win32'
  const arch = process.arch;   // 'x64', 'arm64'
  
  const platformMap: Record<string, Record<string, string>> = {
    'darwin': { 'x64': 'darwin-x86_64', 'arm64': 'darwin-aarch64' },
    'linux':  { 'x64': 'linux-x86_64',  'arm64': 'linux-aarch64' },
    'win32':  { 'x64': 'windows-x86_64', 'arm64': 'windows-aarch64' }
  };
  
  return platformMap[os]?.[arch] ?? 'linux-x86_64';
}
```

### Filter Available Agents

```typescript
function getCompatibleAgents(registry: Registry): Agent[] {
  const platform = getCurrentPlatform();
  
  return registry.agents.filter(agent => {
    // Must have at least one distribution method
    if (!agent.distribution) return false;
    
    // Check if binary available for current platform
    if (agent.distribution.binary?.[platform]) return true;
    
    // npx works on all platforms
    if (agent.distribution.npx) return true;
    
    // uvx works on all platforms
    if (agent.distribution.uvx) return true;
    
    return false;
  });
}
```

### Filter by Features

```typescript
function filterAgents(agents: Agent[], options: {
  requiresAuth?: boolean;    // Agent must support auth
  openSource?: boolean;     // Must be open source
  preferredLicense?: string[];
}): Agent[] {
  return agents.filter(agent => {
    if (options.openSource && !isOpenSource(agent.license)) {
      return false;
    }
    
    if (options.preferredLicense?.length) {
      if (!options.preferredLicense.includes(agent.license)) {
        return false;
      }
    }
    
    return true;
  });
}

function isOpenSource(license: string): boolean {
  const osl = ['MIT', 'Apache-2.0', 'BSD-2-Clause', 'BSD-3-Clause', 'GPL-3.0', 'AGPL-3.0'];
  return osl.some(l => license.toLowerCase().includes(l.toLowerCase()));
}
```

---

## Step 3: Present to User

### Build Agent Display Model

```typescript
interface AgentDisplayItem {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  license: string;
  iconUrl: string;
  installCommand: string;
  distributionType: 'binary' | 'npx' | 'uvx';
}

function buildDisplayItem(agent: Agent, platform: string): AgentDisplayItem {
  const distType = getPreferredDistribution(agent, platform);
  const installCmd = buildInstallCommand(agent, distType, platform);
  
  return {
    id: agent.id,
    name: agent.name,
    version: agent.version,
    description: agent.description,
    author: agent.authors[0] ?? 'Unknown',
    license: agent.license,
    iconUrl: agent.icon ?? '',
    installCommand: installCmd,
    distributionType: distType
  };
}

function getPreferredDistribution(agent: Agent, platform: string): 'binary' | 'npx' | 'uvx' {
  if (agent.distribution.binary?.[platform]) return 'binary';
  if (agent.distribution.npx) return 'npx';
  if (agent.distribution.uvx) return 'uvx';
  throw new Error('No compatible distribution');
}

function buildInstallCommand(agent: Agent, type: 'binary' | 'npx' | 'uvx', platform: string): string {
  switch (type) {
    case 'binary': {
      const target = agent.distribution.binary[platform];
      return `Download from ${target.archive}`;
    }
    case 'npx': {
      const { package: pkg, args = [] } = agent.distribution.npx;
      return `npx ${pkg} ${args.join(' ')}`;
    }
    case 'uvx': {
      const { package: pkg, args = [] } = agent.distribution.uvx;
      return `uvx ${pkg} ${args.join(' ')}`;
    }
  }
}
```

### UI Example (React-like)

```tsx
function AgentSelector({ agents, onSelect }: {
  agents: AgentDisplayItem[],
  onSelect: (agent: AgentDisplayItem) => void
}) {
  return (
    <div className="agent-list">
      {agents.map(agent => (
        <div 
          key={agent.id} 
          className="agent-card"
          onClick={() => onSelect(agent)}
        >
          {agent.iconUrl && <img src={agent.iconUrl} alt="" />}
          <h3>{agent.name}</h3>
          <p>{agent.description}</p>
          <span className="version">v{agent.version}</span>
          <span className="license">{agent.license}</span>
        </div>
      ))}
    </div>
  );
}
```

---

## Step 4: Install the Agent

### For Binary Distribution

```typescript
import { createWriteStream } from 'fs';
import { pipeline } from 'stream/promises';
import * as tar from 'tar'; // or use adm-zip for .zip

async function installBinary(agent: Agent, platform: string, targetDir: string): Promise<string> {
  const target = agent.distribution.binary[platform];
  const archiveUrl = target.archive;
  const cmd = target.cmd;
  const args = target.args ?? [];
  
  // Determine file extension
  const isZip = archiveUrl.endsWith('.zip');
  const extractDir = `${targetDir}/${agent.id}`;
  
  // Download
  const response = await fetch(archiveUrl);
  if (!response.ok) throw new Error(`Download failed: ${response.status}`);
  
  // Extract
  if (isZip) {
    // Use adm-zip or similar
    await extractZip(response.body, extractDir);
  } else {
    await tar.extract({ file: response.body, cwd: extractDir });
  }
  
  // Return full path to executable
  return `${extractDir}/${cmd}`;
}

async function extractZip(stream: ReadableStream, targetDir: string) {
  // Implementation depends on your runtime
}
```

### For npx Distribution

```typescript
async function installNpx(agent: Agent): Promise<string[]> {
  // npx doesn't require pre-installation
  // Just return the command parts
  const { package: pkg, args = [] } = agent.distribution.npx;
  return [pkg, ...args];
}
```

### For uvx Distribution

```typescript
async function installUvx(agent: Agent): Promise<string[]> {
  // uvx doesn't require pre-installation
  const { package: pkg, args = [] } = agent.distribution.uvx;
  return [pkg, ...args];
}
```

---

## Step 5: Launch as ACP Subprocess

### Process Launcher

```typescript
import { spawn, ChildProcess } from 'child_process';

interface AcpClientOptions {
  agentPath: string;        // Path to binary or package name
  distributionType: 'binary' | 'npx' | 'uvx';
  args?: string[];
  env?: Record<string, string>;
  cwd?: string;
}

class AcpProcess {
  private process: ChildProcess | null = null;
  private messageId = 0;
  
  async start(options: AcpClientOptions): Promise<void> {
    const { agentPath, distributionType, args = [], env = {}, cwd } = options;
    
    let command: string;
    let spawnArgs: string[];
    
    switch (distributionType) {
      case 'binary':
        command = agentPath;
        spawnArgs = args;
        break;
      case 'npx':
        command = 'npx';
        spawnArgs = [agentPath, ...args];
        break;
      case 'uvx':
        command = 'uvx';
        spawnArgs = [agentPath, ...args];
        break;
    }
    
    this.process = spawn(command, spawnArgs, {
      stdio: ['pipe', 'pipe', 'pipe'],
      env: { ...process.env, ...env },
      cwd
    });
    
    // Handle stdout (JSON-RPC messages)
    this.process.stdout?.on('data', (data) => {
      const lines = data.toString().split('\n').filter(Boolean);
      for (const line of lines) {
        this.handleMessage(line);
      }
    });
    
    // Handle stderr (logs)
    this.process.stderr?.on('data', (data) => {
      console.error('[Agent]', data.toString());
    });
  }
  
  private handleMessage(line: string) {
    try {
      const msg = JSON.parse(line);
      if (msg.method === 'initialize') {
        this.sendInitialize();
      }
    } catch (e) {
      console.error('Failed to parse message:', e);
    }
  }
  
  private sendInitialize() {
    this.send({
      jsonrpc: '2.0',
      id: 0,
      method: 'initialize',
      params: {
        protocolVersion: 1,
        clientCapabilities: {
          fs: { readTextFile: true, writeTextFile: true },
          terminal: true,
          auth: { terminal: true }
        },
        clientInfo: { name: 'my-client', version: '1.0.0' }
      }
    });
  }
  
  private send(msg: object) {
    this.process?.stdin?.write(JSON.stringify(msg) + '\n');
  }
  
  async stop(): Promise<void> {
    this.process?.kill();
    this.process = null;
  }
}
```

---

## Step 6: Complete ACP Handshake

### Initialize and Session Flow

```typescript
async function connectToAgent(agent: AgentDisplayItem): Promise<AcpSession> {
  const platform = getCurrentPlatform();
  
  // Install agent (if binary)
  let agentPath: string;
  let distType = getAgentDistributionType(agent, platform);
  
  if (distType === 'binary') {
    agentPath = await installBinaryAgent(agent, platform);
  } else if (distType === 'npx') {
    agentPath = `${agent.installCommand}`; // Parse appropriately
  } else {
    agentPath = `${agent.installCommand}`;
  }
  
  // Launch process
  const acp = new AcpProcess();
  await acp.start({
    agentPath,
    distributionType: distType,
    args: distType === 'binary' ? ['acp'] : []
  });
  
  // Wait for initialize response
  const initResponse = await acp.waitForMessage(0);
  
  // Check auth methods
  if (initResponse.result.authMethods?.length > 0) {
    // Handle authentication before creating session
    await handleAuthentication(acp, initResponse.result.authMethods);
  }
  
  // Create session
  const sessionResponse = await acp.sendRequest('session/new', {
    cwd: process.cwd(),
    mcpServers: []
  });
  
  return new AcpSession(acp, sessionResult.sessionId);
}

async function handleAuthentication(acp: AcpProcess, authMethods: AuthMethod[]) {
  // Prefer agent auth (default)
  const agentAuth = authMethods.find(m => m.type === 'agent' || !m.type);
  
  if (agentAuth) {
    await acp.sendRequest('authenticate', { methodId: agentAuth.id });
  }
}
```

---

## Complete Example: Simple CLI Client

```typescript
#!/usr/bin/env node

import { fetchRegistry, getCompatibleAgents } from './registry-client';
import { installBinaryAgent } from './installer';
import { AcpProcess } from './acp-process';

async function main() {
  console.log('Fetching ACP registry...');
  const registry = await fetchRegistry();
  
  const platform = getCurrentPlatform();
  const agents = getCompatibleAgents(registry, platform);
  
  console.log(`\nAvailable agents for ${platform}:`);
  agents.forEach((a, i) => {
    console.log(`  ${i + 1}. ${a.name} (v${a.version}) - ${a.description}`);
  });
  
  // For demo: use first agent
  const selected = agents[0];
  console.log(`\nInstalling ${selected.name}...`);
  
  const agentPath = await installBinaryAgent(selected, platform);
  console.log(`Launching ${selected.name}...`);
  
  const acp = new AcpProcess();
  await acp.start({
    agentPath,
    distributionType: 'binary',
    args: ['acp']
  });
  
  console.log('Agent running. Press Ctrl+C to exit.');
  
  // Keep process running
  process.on('SIGINT', async () => {
    await acp.stop();
    process.exit(0);
  });
}

main().catch(console.error);
```

---

## Caching Strategy

For production clients, implement caching:

```typescript
class RegistryCache {
  private cache: Registry | null = null;
  private lastFetch = 0;
  private readonly TTL = 60 * 60 * 1000; // 1 hour
  
  async get(): Promise<Registry> {
    const now = Date.now();
    
    if (this.cache && (now - this.lastFetch) < this.TTL) {
      return this.cache;
    }
    
    this.cache = await fetchRegistry();
    this.lastFetch = now;
    return this.cache;
  }
  
  invalidate(): void {
    this.cache = null;
    this.lastFetch = 0;
  }
}
```

---

## Error Handling

```typescript
try {
  const registry = await fetchRegistry();
} catch (e) {
  if (e instanceof TypeError && e.message.includes('fetch')) {
    // Network error - use cached or default
    console.warn('Network error, using cached registry');
    return getCachedRegistry();
  }
  throw e;
}

// Handle agent-specific errors
async function withAgentErrorHandling(fn: () => Promise<void>) {
  try {
    await fn();
  } catch (e) {
    if (e instanceof Error) {
      if (e.message.includes('auth_required')) {
        throw new Error('Agent requires authentication. Please authenticate first.');
      }
      if (e.message.includes('protocol_version')) {
        throw new Error('Agent protocol version mismatch.');
      }
    }
    throw e;
  }
}
```

---

## Best Practices

1. **Cache the registry** - Don't fetch on every launch (TTL: 1 hour)
2. **Verify downloads** - Check checksums for binary archives
3. **Handle auth gracefully** - Present auth UI when needed
4. **Clean up on exit** - Kill agent process properly
5. **Log verbosely** - Capture stderr for debugging
6. **Version pinned** - Pin to specific registry version for stability
7. **Fallback** - Have default agents if registry unavailable