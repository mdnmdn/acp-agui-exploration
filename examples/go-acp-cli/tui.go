package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coder/acp-go-sdk"
)

// Styles
var (
	appStyle = lipgloss.NewStyle().Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, false, true).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginRight(1)

	activeTabStyle = tabStyle.
			BorderForeground(lipgloss.Color("#FF75B7")).
			Background(lipgloss.Color("#FF75B7")).
			Foreground(lipgloss.Color("#FAFAFA"))

	chatBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF75B7")).
			Padding(0, 1)

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262")).
			Padding(0, 1)

	inputContainerStyle = lipgloss.NewStyle().
				Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF75B7"))

	// Message styles
	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ADD8"))

	agentMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	systemMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)
)

type state int

const (
	stateSelectingAgent state = iota
	stateSelectingModel
	stateChatting
)

type message struct {
	tabIndex int
	sender   string
	content  string
	isLog    bool
}

type tab struct {
	selected   agentInfo
	modelId    acp.ModelId
	sessionId  acp.SessionId
	acpClient  *acp.ClientSideConnection
	agentCmd   *exec.Cmd
	messages   []message
	logs       []string
	viewport   viewport.Model
	logPort    viewport.Model
	input      textinput.Model
	acpMu      sync.Mutex
}

type model struct {
	state      state
	agentList  list.Model
	modelList  list.Model
	tabs       []*tab
	activeTab  int
	showLogs   bool
	width      int
	height     int
	err        error
	msgChan    chan message
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.killAll()
			return m, tea.Quit
		case "q":
			if m.state == stateSelectingAgent || m.state == stateSelectingModel {
				if len(m.tabs) == 0 {
					return m, tea.Quit
				}
				m.state = stateChatting
				return m, nil
			}
		case "enter":
			switch m.state {
			case stateSelectingAgent:
				item := m.agentList.SelectedItem()
				if item != nil {
					selectedAgent := item.(agentInfo)
					m.state = stateChatting
					return m, m.startAgent(selectedAgent)
				}
			case stateSelectingModel:
				item := m.modelList.SelectedItem()
				if item != nil && len(m.tabs) > 0 {
					t := m.tabs[m.activeTab]
					t.modelId = acp.ModelId(item.(modelInfo).id)
					m.state = stateChatting
					return m, m.setAgentModel(m.activeTab)
				}
			default:
				if len(m.tabs) > 0 {
					t := m.tabs[m.activeTab]
					content := t.input.Value()
					if content != "" {
						t.input.SetValue("")
						t.messages = append(t.messages, message{tabIndex: m.activeTab, sender: "User", content: content})
						m.updateViewport(m.activeTab)
						return m, m.sendACPMessage(m.activeTab, content)
					}
				}
			}
		case "ctrl+l":
			m.showLogs = !m.showLogs
			m.updateLayout()
			return m, nil
		case "ctrl+m":
			if m.state == stateChatting && len(m.tabs) > 0 {
				m.state = stateSelectingModel
				return m, nil
			}
		case "ctrl+a":
			m.state = stateSelectingAgent
			return m, nil
		case "ctrl+w":
			if len(m.tabs) > 0 {
				m.closeTab(m.activeTab)
				if len(m.tabs) == 0 {
					m.state = stateSelectingAgent
				} else if m.activeTab >= len(m.tabs) {
					m.activeTab = len(m.tabs) - 1
				}
				m.updateLayout()
				return m, nil
			}
		case "tab":
			if len(m.tabs) > 1 {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				m.updateLayout()
				return m, nil
			}
		case "shift+tab":
			if len(m.tabs) > 1 {
				m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
				m.updateLayout()
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case message:
		if msg.tabIndex < len(m.tabs) {
			t := m.tabs[msg.tabIndex]
			if msg.isLog {
				t.logs = append(t.logs, msg.content)
				t.logPort.SetContent(strings.Join(t.logs, "\n"))
				t.logPort.GotoBottom()
			} else {
				t.messages = append(t.messages, msg)
				m.updateViewport(msg.tabIndex)
			}
		}
		return m, m.waitForMsg()

	case errMsg:
		m.err = msg.err
		if len(m.tabs) > 0 {
			t := m.tabs[m.activeTab]
			t.messages = append(t.messages, message{tabIndex: m.activeTab, sender: "System", content: fmt.Sprintf("Error: %v", msg.err)})
			m.updateViewport(m.activeTab)
		}
		return m, m.waitForMsg()
	
	case successMsg:
		return m, nil
	}

	switch m.state {
	case stateSelectingAgent:
		m.agentList, cmd = m.agentList.Update(msg)
	case stateSelectingModel:
		m.modelList, cmd = m.modelList.Update(msg)
	case stateChatting:
		if len(m.tabs) > 0 {
			t := m.tabs[m.activeTab]
			t.input, cmd = t.input.Update(msg)
			var vpCmd tea.Cmd
			t.viewport, vpCmd = t.viewport.Update(msg)
			cmd = tea.Batch(cmd, vpCmd)
		}
	}

	return m, cmd
}

func (m *model) killAll() {
	for _, t := range m.tabs {
		if t.agentCmd != nil && t.agentCmd.Process != nil {
			_ = t.agentCmd.Process.Kill()
		}
	}
}

func (m *model) closeTab(index int) {
	t := m.tabs[index]
	if t.agentCmd != nil && t.agentCmd.Process != nil {
		_ = t.agentCmd.Process.Kill()
	}
	m.tabs = append(m.tabs[:index], m.tabs[index+1:]...)
}

func (m *model) updateLayout() {
	chatWidth := m.width - 2
	if m.showLogs {
		chatWidth = (m.width * 2 / 3) - 2
	}

	m.agentList.SetSize(m.width-4, m.height-4)
	m.modelList.SetSize(m.width-4, m.height-4)

	for i, t := range m.tabs {
		t.viewport.Width = chatWidth - 6
		t.viewport.Height = m.height - 8
		t.logPort.Width = m.width - chatWidth - 4
		t.logPort.Height = m.height - 4
		t.input.Width = chatWidth - 8
		m.updateViewport(i)
	}
}

func (m *model) updateViewport(tabIndex int) {
	t := m.tabs[tabIndex]
	var sb strings.Builder
	maxBubbleWidth := int(float64(t.viewport.Width) * 0.85)

	for _, msg := range t.messages {
		if msg.sender == "System" {
			sb.WriteString(systemMsgStyle.Render("System: " + msg.content))
			sb.WriteString("\n")
			continue
		}

		if msg.sender == "User" {
			// User on the LEFT
			// Calculate bubble width based on longest line
			lines := strings.Split(msg.content, "\n")
			longestLine := 0
			for _, l := range lines {
				if len(l) > longestLine {
					longestLine = len(l)
				}
			}
			bubbleWidth := longestLine
			if bubbleWidth > maxBubbleWidth {
				bubbleWidth = maxBubbleWidth
			}

			// Wrap and justify right within the bubble
			wrapped := lipgloss.NewStyle().Width(bubbleWidth).Align(lipgloss.Right).Render(msg.content)
			line := fmt.Sprintf("[U] %s", wrapped)
			sb.WriteString(userMsgStyle.Render(line))
		} else {
			// Agent on the RIGHT
			bubbleWidth := maxBubbleWidth
			wrapped := lipgloss.NewStyle().Width(bubbleWidth).Render(msg.content)
			
			// Full row width is viewport width. Text is right-aligned.
			row := lipgloss.NewStyle().Width(t.viewport.Width).Align(lipgloss.Right).Render(fmt.Sprintf("%s [A]", wrapped))
			sb.WriteString(agentMsgStyle.Render(row))
		}
		sb.WriteString("\n\n")
	}
	t.viewport.SetContent(sb.String())
	t.viewport.GotoBottom()
}

func (m *model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress Ctrl+C to quit", m.err)
	}

	if m.state == stateSelectingAgent {
		return appStyle.Render(m.agentList.View())
	}

	if m.state == stateSelectingModel {
		return appStyle.Render(m.modelList.View())
	}

	if len(m.tabs) == 0 {
		return appStyle.Render("No active sessions. Press Ctrl+A to select an agent.")
	}

	activeTab := m.tabs[m.activeTab]

	// Tab bar
	var tabs []string
	for i, t := range m.tabs {
		style := tabStyle
		if i == m.activeTab {
			style = activeTabStyle
		}
		tabs = append(tabs, style.Render(t.selected.name))
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Chat View
	modelDisplay := string(activeTab.modelId)
	if modelDisplay == "" {
		modelDisplay = "None"
	}

	// Chat Frame Border with Model name
	modelLabel := " Model: " + modelDisplay + " "
	borderWidth := activeTab.viewport.Width + 2
	topBorder := "╭" + strings.Repeat("─", borderWidth-len(modelLabel)) + modelLabel + "╮"
	
	// Bottom help
	help := helpStyle.Width(m.width).Render("TAB: next • Ctrl+W: close • Ctrl+A: new agent • Ctrl+M: model • Ctrl+L: logs • Ctrl+C: quit")

	chatContent := chatBoxStyle.
		BorderTop(false).
		Width(activeTab.viewport.Width + 2).
		Height(activeTab.viewport.Height + 1).
		Render(activeTab.viewport.View())

	chatView := lipgloss.JoinVertical(lipgloss.Left,
		tabBar,
		topBorder,
		chatContent,
		inputContainerStyle.Width(activeTab.viewport.Width+2).Render(inputStyle.Render("> ")+activeTab.input.View()),
	)

	fullView := chatView
	if m.showLogs {
		logView := logBoxStyle.Width(activeTab.logPort.Width+2).Height(activeTab.logPort.Height+2).Render(activeTab.logPort.View())
		fullView = lipgloss.JoinHorizontal(lipgloss.Top, chatView, logView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, fullView, help)
}
