package main

import (
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Agent definition
type agentInfo struct {
	name    string
	command string
	args    []string
	desc    string
}

func (a agentInfo) Title() string       { return a.name }
func (a agentInfo) Description() string { return a.desc }
func (a agentInfo) FilterValue() string { return a.name }

var agents = []list.Item{
	agentInfo{name: "claude", command: "claude", args: []string{"--acp"}, desc: "Anthropic Claude Agent"},
	agentInfo{name: "gemini", command: "gemini", args: []string{"--acp"}, desc: "Google Gemini Agent"},
	agentInfo{name: "opencode", command: "opencode", args: []string{"--acp"}, desc: "OpenCode Agent"},
	agentInfo{name: "vibe", command: "vibe", args: []string{"--acp"}, desc: "Vibe Agent"},
	agentInfo{name: "codex", command: "openai-codex", args: []string{"--acp"}, desc: "OpenAI Codex Agent"},
	agentInfo{name: "mock", command: os.Args[0], args: []string{"mock-agent"}, desc: "Internal Mock ACP Agent"},
}

// Model definition
type modelInfo struct {
	id   string
	desc string
}

func (m modelInfo) Title() string       { return m.id }
func (m modelInfo) Description() string { return m.desc }
func (m modelInfo) FilterValue() string { return m.id }

var defaultModels = []list.Item{
	modelInfo{id: "claude-3-5-sonnet", desc: "Anthropic Claude 3.5 Sonnet"},
	modelInfo{id: "claude-3-opus", desc: "Anthropic Claude 3 Opus"},
	modelInfo{id: "gemini-1.5-pro", desc: "Google Gemini 1.5 Pro"},
	modelInfo{id: "gpt-4o", desc: "OpenAI GPT-4o"},
	modelInfo{id: "deepseek-coder", desc: "DeepSeek Coder V2"},
}

func main() {
	Execute()
}

func startTui() {
	al := list.New(agents, list.NewDefaultDelegate(), 0, 0)
	al.Title = "Select ACP Agent"

	ml := list.New(defaultModels, list.NewDefaultDelegate(), 0, 0)
	ml.Title = "Select Model"

	m := model{
		state:     stateSelectingAgent,
		agentList: al,
		modelList: ml,
		msgChan:   make(chan message, 100),
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

type errMsg struct{ err error }
type successMsg struct {
	tabIndex int
}
