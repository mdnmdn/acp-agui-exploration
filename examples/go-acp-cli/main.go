package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
)

// agentInfo describes an ACP-compatible agent binary.
// It is shared between the ACP layer (acp.go) and the UI layer (tui.go).
type agentInfo struct {
	name    string
	command string
	args    []string
	desc    string
}

func (a agentInfo) Title() string       { return a.name }
func (a agentInfo) Description() string { return a.desc }
func (a agentInfo) FilterValue() string { return a.name }

// modelInfo describes a selectable model.
type modelInfo struct {
	id   string
	desc string
}

func (m modelInfo) Title() string       { return m.id }
func (m modelInfo) Description() string { return m.desc }
func (m modelInfo) FilterValue() string { return m.id }

var defaultAgents = []list.Item{
	agentInfo{name: "claude", command: "claude", args: []string{"--acp"}, desc: "Anthropic Claude"},
	agentInfo{name: "gemini", command: "gemini", args: []string{"--acp"}, desc: "Google Gemini"},
	agentInfo{name: "opencode", command: "opencode", args: []string{"--acp"}, desc: "OpenCode"},
	agentInfo{name: "codex", command: "openai-codex", args: []string{"--acp"}, desc: "OpenAI Codex"},
	agentInfo{name: "mock", command: os.Args[0], args: []string{"mock-agent"}, desc: "Built-in mock agent (echo)"},
}

// agents is the live list, populated from the registry at startup.
var agents = defaultAgents

var defaultModels = []list.Item{
	modelInfo{id: "claude-sonnet-4-6", desc: "Anthropic Claude Sonnet 4.6"},
	modelInfo{id: "claude-opus-4-6", desc: "Anthropic Claude Opus 4.6"},
	modelInfo{id: "gemini-1.5-pro", desc: "Google Gemini 1.5 Pro"},
	modelInfo{id: "gpt-4o", desc: "OpenAI GPT-4o"},
	modelInfo{id: "deepseek-coder", desc: "DeepSeek Coder V2"},
}

func main() {
	loadRegistryAgents()
	Execute()
}

func loadRegistryAgents() {
	loaded, err := LoadAgentsFromRegistry()
	if err != nil {
		fmt.Printf("Warning: registry unavailable (%v), using built-ins\n", err)
		return
	}
	agents = loaded
	fmt.Printf("Loaded %d agents from ACP registry\n", len(agents))
}
