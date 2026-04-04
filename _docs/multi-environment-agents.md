# Multi-Environment Agent Isolation

> Analysis of strategies to run multiple isolated instances of AI agent CLIs (claude, gemini, opencode, mistral-vibe) with different authentications and settings

## Overview

This document analyzes how to "isolate" major AI agent CLIs to run separate instances with different authentications, settings, or configurations. This is useful for:
- Using personal vs work accounts
- Managing multiple organization credentials
- Running different model configurations
- Separating MCP server configurations
- Testing different agent settings

---

## Claude Code (claude)

### Isolation Mechanism

Claude Code supports **native isolation** via the `CLAUDE_CONFIG_DIR` environment variable.

| Aspect | Details |
|--------|---------|
| Config Directory | Default: `~/.claude` |
| Isolation Env Var | `CLAUDE_CONFIG_DIR` |
| Status | **Fully Supported** |

### Configuration Files Location

| File | Path | Purpose |
|------|------|---------|
| Auth Tokens | `~/.claude/settings.json` | API keys and authentication |
| MCP Servers | `~/.claude/mcp_servers.json` | MCP server configurations |
| Skills | `~/.claude/skills/` | Custom skills |
| Conversations | `~/.claude/conversations/` | Chat history |
| Keybindings | `~/.claude/keybindings.json` | Keyboard shortcuts |
| Logs | `~/.claude/logs/` | Session logs |

### Full Directory Structure

```
~/.claude/
├── settings.json       # User settings and auth tokens
├── mcp_servers.json    # MCP server configurations
├── keybindings.json   # Keyboard shortcuts
├── skills/            # Custom skills
├── conversations/     # Chat history
└── logs/              # Session logs
```

### Usage Examples

```bash
# Default instance (personal)
claude

# Secondary instance (work)
CLAUDE_CONFIG_DIR=~/.claude-work claude

# Using shell aliases
alias cc='claude'
alias cc-work='CLAUDE_CONFIG_DIR=~/.claude-work claude'
alias cc-client-a='CLAUDE_CONFIG_DIR=~/.claude-client-a claude'
```

### Sharing Configuration via Symlinks

```bash
# Share MCP servers across instances
ln -s ~/.claude/mcp_servers.json ~/.claude-work/mcp_servers.json

# Share skills
ln -s ~/.claude/skills ~/.claude-work/skills

# WARNING: Don't share settings.json (contains auth tokens) unless using same account
```

### Notes

- Each instance maintains separate conversation history
- Authentication tokens in `settings.json` are isolated per config directory
- Multiple instances can run simultaneously without conflicts

---

## OpenCode

### Isolation Mechanism

OpenCode has **partial support** via `OPENCODE_CONFIG_DIR`, but has known issues with config isolation.

| Aspect | Details |
|--------|---------|
| Config Directory | Default: `~/.config/opencode` |
| Isolation Env Var | `OPENCODE_CONFIG_DIR` (since PR #629) |
| Status | **Partial - Has Known Bugs** |

### Configuration Files Location

| File | Path | Purpose |
|------|------|---------|
| Provider Settings | `~/.config/opencode/settings.json` | Provider configs and auth |
| MCP Servers | `~/.config/opencode/mcp_servers.json` | MCP configurations |
| Plugins | `~/.config/opencode/plugins/` | Plugin configurations |
| Cache | `~/.config/opencode/cache/` | Session cache |

### Full Directory Structure

```
~/.config/opencode/
├── settings.json       # Provider settings and auth
├── mcp_servers.json    # MCP configurations
├── plugins/            # Plugin configurations
└── cache/              # Session cache
```

### Known Issues

1. **MCP servers from ~/.claude are loaded even when OPENCODE_CONFIG_DIR is set** (#18691)
   - OpenCode incorrectly loads Claude Code's MCP config
   - Breaks isolation for MCP servers

2. **Plugin config files deleted when multiple instances run concurrently** (#16450)
   - Race condition when writing to shared config directory

3. **Feature Request: Multiple Auth Profiles** (#16866)
   - Request for named auth profiles per provider (e.g., `anthropic/work`, `anthropic/personal`)
   - Not yet implemented

### Usage (When Working)

```bash
# Set custom config directory
OPENCODE_CONFIG_DIR=~/.opencode-work opencode
```

### Workaround Strategies

1. **Use wrapper scripts** to manage environment variables
2. **Avoid running multiple OpenCode instances concurrently** to prevent config corruption
3. **Monitor issue #18691** for fix status

---

## Gemini CLI (gemini)

### Isolation Mechanism

Gemini CLI does **not** have native config directory isolation. It uses a **layered configuration system** instead.

| Aspect | Details |
|--------|---------|
| Config Directory | Default: `~/.gemini` |
| Isolation Env Var | **None available** |
| Status | **Not Supported** |

### Configuration Files Location

| File | Path | Purpose |
|------|------|---------|
| Settings | `~/.gemini/settings.json` | User preferences and auth |
| System Defaults | `/etc/gemini-cli/system-defaults.json` | System-wide defaults (Linux) |
| System Settings | `/etc/gemini-cli/settings.json` | System-wide overrides |
| Project Settings | `.gemini/settings.json` | Project-specific config |

### Full Directory Structure

```
~/.gemini/
├── settings.json       # User settings
├── .env               # API keys (auto-loaded)
└── tmp/               # Temporary files (per-project)
    └── <project_hash>/
        └── shell_history
```

### Configuration Layers (Precedence Order)

1. Default values (hardcoded)
2. System defaults (`/etc/gemini-cli/system-defaults.json`)
3. User settings (`~/.gemini/settings.json`)
4. Project settings (`.gemini/settings.json`)
5. System settings (`/etc/gemini-cli/settings.json`)
6. Environment variables
7. Command-line arguments

### Environment Variables for Auth

```bash
# API Key based authentication
export GEMINI_API_KEY="your-api-key"

# Google Cloud authentication
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
export GOOGLE_CLOUD_PROJECT="your-project"
export GOOGLE_CLOUD_LOCATION="us-central1"

# Vertex AI
export GOOGLE_API_KEY="your-google-cloud-key"
```

### Workaround Strategies

1. **Use different user accounts** - Run gemini CLI under different system users (complex)

2. **Project-level isolation** - Use `.gemini/settings.json` in project directories:
   ```json
   {
     "model": { "name": "gemini-1.5-pro" },
     "security": { "auth": { "selectedType": "apiKey" } }
   }
   ```

3. **Environment variable prefixes** - Gemini doesn't support this natively

4. **Feature Request #2449** - Profile support requested but closed as not planned

### Limitations

- Cannot run multiple simultaneous instances with different authentications
- No native support for separate config directories
- All auth is user-level, not profile-based

---

## Mistral Vibe (vibe)

### Isolation Mechanism

Mistral Vibe supports **full isolation** via the `VIBE_HOME` environment variable.

| Aspect | Details |
|--------|---------|
| Config Directory | Default: `~/.vibe` |
| Isolation Env Var | `VIBE_HOME` |
| Status | **Fully Supported** |

### Configuration Files Location

| File | Path | Purpose |
|------|------|---------|
| API Keys | `~/.vibe/.env` | Authentication tokens |
| Main Config | `~/.vibe/config.toml` | Agent and model configuration |
| Custom Agents | `~/.vibe/agents/` | Custom agent configurations |
| Custom Prompts | `~/.vibe/prompts/` | Custom system prompts |
| Custom Tools | `~/.vibe/tools/` | Custom tools |
| Trusted Folders | `~/.vibe/trusted_folders.toml` | Folder trust settings |
| Logs | `~/.vibe/logs/` | Session logs |

### Full Directory Structure

```
~/.vibe/
├── .env                    # API keys (MISTRAL_API_KEY, etc.)
├── config.toml             # Main configuration
├── agents/                 # Custom agent configurations (*.toml)
├── prompts/                # Custom system prompts (*.md)
├── tools/                  # Custom tools
├── trusted_folders.toml    # Trusted folders list
└── logs/                   # Session logs
```

### Configuration Search Order

1. `./.vibe/config.toml` (project directory)
2. `~/.vibe/config.toml` (user home directory)

### API Key Authentication Methods

```bash
# Method 1: Interactive setup (first run)
vibe  # Will prompt for API key

# Method 2: Environment variable
export MISTRAL_API_KEY="your-mistral-api-key"

# Method 3: .env file (~/.vibe/.env)
MISTRAL_API_KEY=your-api-key
```

### Usage Examples

```bash
# Default instance
vibe

# Secondary instance (work) - set VIBE_HOME
VIBE_HOME=~/.vibe-work vibe

# Using shell aliases
alias vibe-personal='VIBE_HOME=~/.vibe-personal vibe'
alias vibe-work='VIBE_HOME=~/.vibe-work vibe'
```

### Custom Agent Configurations

You can create multiple agent profiles in `~/.vibe/agents/`:

```toml
# ~/.vibe/agents/work.toml
active_model = "devstral-2"
system_prompt_id = "work-context"
disabled_tools = ["search_replace"]
```

```bash
# Use specific agent
vibe --agent work
```

### Provider and Model Configuration

Vibe supports multiple providers in `config.toml`:

```toml
[[providers]]
name = "openrouter"
api_base = "https://openrouter.ai/api/v1"
api_key_env_var = "OPENROUTER_API_KEY"

[[models]]
name = "anthropic/claude-3-sonnet"
provider = "openrouter"
alias = "claude-work"
```

### Notes

- Each VIBE_HOME directory has independent `.env` with separate API keys
- Supports custom agents, prompts, and tools per configuration
- Full isolation allows running multiple instances simultaneously

---

## Comparison Summary

| Feature | Claude Code | OpenCode | Gemini CLI | Mistral Vibe |
|---------|-------------|----------|------------|--------------|
| Config Dir Override | ✅ `CLAUDE_CONFIG_DIR` | ⚠️ `OPENCODE_CONFIG_DIR` (buggy) | ❌ Not available | ✅ `VIBE_HOME` |
| Multi-instance Support | ✅ Full | ⚠️ Partial (bugs) | ❌ Not supported | ✅ Full |
| Auth Profiles | ✅ Via separate configs | 🔄 Requested (#16866) | ❌ Not available | ✅ Via agents |
| MCP Isolation | ✅ Via config dir | ❌ Broken (#18691) | ✅ Via project config | ✅ Via config |
| Simultaneous Run | ✅ Supported | ⚠️ Risky | ❌ Not supported | ✅ Supported |

---

## Recommended Solutions

### For Claude Code

**Use native `CLAUDE_CONFIG_DIR`** - Full support, recommended for production.

```bash
# Shell setup (~/.zshrc)
export CLAUDE_CONFIG_DIR=~/.claude-personal
alias cc-personal='CLAUDE_CONFIG_DIR=~/.claude-personal claude'
alias cc-work='CLAUDE_CONFIG_DIR=~/.claude-work claude'
alias cc='claude'  # Uses default ~/.claude
```

### For OpenCode

**Use `OPENCODE_CONFIG_DIR` with caution** - Monitor for bug fixes.

```bash
# Monitor known issues:
# - #18691: MCP servers from ~/.claude loaded incorrectly
# - #16450: Config corruption with concurrent instances

OPENCODE_CONFIG_DIR=~/.opencode-work opencode
```

### For Gemini CLI

**Use project-level config or environment variables** - Limited but functional.

```bash
# Project-specific config: ./project/.gemini/settings.json
{
  "model": { "name": "gemini-2.0-flash" },
  "mcpServers": { "project-tools": { "command": "..." } }
}

# For different auth, use environment variables per project:
# In project .env file
GEMINI_API_KEY=project-specific-key
```

### For Mistral Vibe

**Use native `VIBE_HOME`** - Full support, recommended for production.

```bash
# Shell setup (~/.zshrc)
alias vibe-personal='VIBE_HOME=~/.vibe-personal vibe'
alias vibe-work='VIBE_HOME=~/.vibe-work vibe'

# Create separate configs
mkdir ~/.vibe-work
cp ~/.vibe/config.toml ~/.vibe-work/
# Edit ~/.vibe-work/config.toml with work-specific settings
# Add work API key in ~/.vibe-work/.env
```

---

## Future Considerations

### OpenCode
- Track PR #629 for `OPENCODE_CONFIG_DIR` improvements
- Monitor #18691 (MCP isolation bug)
- Track #16866 (multi-auth profiles feature request)

### Gemini CLI
- No current roadmap for config directory isolation
- Alternative: Use MCP to expose different tools per project

### Mistral Vibe
- Monitor issue #409: XDG Base Directory Specification compliance request

---

## References

- Claude Code: [Multiple Instances Guide](https://www.elliotjreed.com/ai/running-multiple-claude-code-instances/)
- OpenCode: [GitHub Issue #16866](https://github.com/anomalyco/opencode/issues/16866) - Multi-auth profiles
- OpenCode: [GitHub Issue #18691](https://github.com/anomalyco/opencode/issues/18691) - MCP config bug
- Gemini CLI: [Configuration Docs](https://google-gemini.github.io/gemini-cli/docs/get-started/configuration.html)
- Gemini CLI: [GitHub Issue #2449](https://github.com/google-gemini/gemini-cli/issues/2449) - Profile support
- Mistral Vibe: [Configuration Docs](https://docs.mistral.ai/mistral-vibe/introduction/configuration)
- Mistral Vibe: [GitHub Issue #409](https://github.com/mistralai/mistral-vibe/issues/409) - XDG specs