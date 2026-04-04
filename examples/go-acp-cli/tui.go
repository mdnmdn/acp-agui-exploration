package main

// tui.go — bubbletea TUI frontend.
//
// Frontier contract: this file may only interact with the ACP layer through
// AgentSession (defined in acp.go). It must not import the ACP SDK directly.
// To swap the UI, replace this file; acp.go and main.go stay untouched.

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	// Chrome
	tabStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, false, true).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginRight(1)

	activeTabStyle = tabStyle.
			BorderForeground(lipgloss.Color("#FF75B7")).
			Background(lipgloss.Color("#2D1040")).
			Foreground(lipgloss.Color("#FF75B7"))

	chatFrameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3D2060"))

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF75B7")).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3A3A3A")).
			Background(lipgloss.Color("#0A0A0A")).
			Padding(0, 1)

	// Chat bubbles
	userBubbleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00ADD8")).
			Background(lipgloss.Color("#001E30")).
			Foreground(lipgloss.Color("#B0E8FF")).
			Padding(0, 1)

	agentBubbleStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Background(lipgloss.Color("#130D28")).
				Foreground(lipgloss.Color("#D8CCFF")).
				Padding(0, 1)

	userLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ADD8")).
			Bold(true)

	agentLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9D76FF")).
			Bold(true)

	systemMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#484848")).
			Italic(true)

	// Thinking bubble (shown in extended mode)
	thoughtBubbleStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#2A2A3A")).
				Foreground(lipgloss.Color("#505060")).
				Italic(true).
				Padding(0, 1)

	thoughtLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3A3A4A"))

	// Tool card (shown in extended mode)
	toolCardBaseStyle = lipgloss.NewStyle().
				Padding(0, 1)

	toolLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666677"))

	// Activity status bar (shown in in_progress mode when agent is busy)
	activityBarStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#0D0D20")).
				Foreground(lipgloss.Color("#7777AA")).
				Italic(true).
				Padding(0, 2)

	// Mode indicator shown in the status line when extended mode is on
	extendedModeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1A0D28")).
				Foreground(lipgloss.Color("#7D56F4")).
				Padding(0, 2)

	// Error toast
	toastStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#8B0000")).
			Foreground(lipgloss.Color("#FFCCCC")).
			Padding(0, 2).
			Bold(true)

	// Modal overlay
	modalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF75B7")).
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF75B7")).
			Bold(true)

	modalHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")).
			Italic(true)

	// Log panel
	logPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2A2A4A"))

	logHeaderStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#13132A")).
			Foreground(lipgloss.Color("#6666AA")).
			Bold(true).
			Padding(0, 1)

	logSendStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true)
	logRecvStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9D76FF")).Bold(true)
	logStderrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6666")).Bold(true)
	logBodyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#555566"))

	// Welcome screen
	welcomeTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	welcomeSubStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444455")).
			Italic(true)
)

// ── Domain types ──────────────────────────────────────────────────────────────

type appState int

const (
	stateSelectingAgent appState = iota
	stateSelectingModel
	stateChatting
)

// thinkMode controls how thinking and tool-use events are displayed.
type thinkMode int

const (
	// thinkModeProgress shows a one-line live status ("thinking…" / "working on X")
	// that disappears once the action completes. Thought/tool entries are hidden
	// from the message history.
	thinkModeProgress thinkMode = iota

	// thinkModeExtended renders full thought bubbles and tool-call cards inline
	// in the conversation, preserving the complete agent activity history.
	thinkModeExtended
)

// msgKind distinguishes the three kinds of messages stored in a tab.
type msgKind int

const (
	kindChat    msgKind = iota // regular user / agent / system text
	kindThought                // agent reasoning chunk
	kindTool                   // tool call card
)

// chatMessage is one entry in the conversation view.
type chatMessage struct {
	kind       msgKind
	sender     string // "User", "System", agent name, or internal tag
	content    string
	toolCallId string // kindTool only
	toolStatus string // kindTool only: pending / in_progress / completed / failed
}

// tab holds per-session UI state.
type tab struct {
	session     *AgentSession
	ready       bool
	thinkMode   thinkMode
	isThinking  bool              // true while receiving thought chunks
	activeTools map[string]string // toolCallId → title, only pending/in_progress entries
	messages    []chatMessage
	logs        []string
	viewport    viewport.Model
	logPort     viewport.Model
	input       textinput.Model
}

// activityStatus returns a one-line description of what the agent is currently
// doing, or "" if it is idle. Used by thinkModeProgress.
func (t *tab) activityStatus() string {
	// Active tool takes priority over generic thinking.
	for _, title := range t.activeTools {
		return "  🔧  " + title
	}
	if t.isThinking {
		return "  💭  thinking…"
	}
	return ""
}

// ── Bubbletea model ───────────────────────────────────────────────────────────

type model struct {
	state            appState
	agentList        list.Model
	modelList        list.Model
	tabs             []*tab
	activeTab        int
	showLogs         bool
	width            int
	height           int
	errorToast       string
	preSelectedAgent string
}

// ── Tea message types ─────────────────────────────────────────────────────────

type sessionEventMsg struct {
	session *AgentSession
	event   AgentEvent
}

type sessionStartedMsg struct {
	session *AgentSession
	err     error
}

type sessionSendErrMsg struct {
	session *AgentSession
	err     error
}

type clearToastMsg struct{}

// ── Commands ──────────────────────────────────────────────────────────────────

func startSessionCmd(s *AgentSession) tea.Cmd {
	return func() tea.Msg {
		cwd, err := os.Getwd()
		if err != nil {
			return sessionStartedMsg{session: s, err: err}
		}
		return sessionStartedMsg{session: s, err: s.Start(context.Background(), cwd)}
	}
}

func listenForEvents(s *AgentSession) tea.Cmd {
	return func() tea.Msg {
		return sessionEventMsg{session: s, event: <-s.Events()}
	}
}

func sendMessageCmd(s *AgentSession, content string) tea.Cmd {
	return func() tea.Msg {
		if err := s.Send(context.Background(), content); err != nil {
			return sessionSendErrMsg{session: s, err: err}
		}
		return nil
	}
}

func setModelCmd(s *AgentSession, modelId string) tea.Cmd {
	return func() tea.Msg {
		if err := s.SetModel(context.Background(), modelId); err != nil {
			return sessionSendErrMsg{session: s, err: err}
		}
		return nil
	}
}

func clearToastAfter(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return clearToastMsg{}
	}
}

// ── model helpers ─────────────────────────────────────────────────────────────

func (m *model) openNewTab(selected agentInfo) tea.Cmd {
	s := NewAgentSession(selected)

	ti := textinput.New()
	ti.Placeholder = "Type a message…"
	ti.Focus()

	t := &tab{
		session:     s,
		thinkMode:   thinkModeProgress,
		activeTools: make(map[string]string),
		messages:    []chatMessage{},
		input:       ti,
		viewport:    viewport.New(0, 0),
		logPort:     viewport.New(0, 0),
	}
	m.tabs = append(m.tabs, t)
	m.activeTab = len(m.tabs) - 1
	m.updateLayout()

	return tea.Batch(startSessionCmd(s), listenForEvents(s))
}

func (m *model) tabForSession(s *AgentSession) (int, *tab) {
	for i, t := range m.tabs {
		if t.session == s {
			return i, t
		}
	}
	return -1, nil
}

func (m *model) showToast(msg string) tea.Cmd {
	m.errorToast = msg
	return clearToastAfter(10 * time.Second)
}

func (m *model) killAll() {
	for _, t := range m.tabs {
		t.session.Close()
	}
}

func (m *model) closeTab(index int) {
	m.tabs[index].session.Close()
	m.tabs = append(m.tabs[:index], m.tabs[index+1:]...)
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m *model) Init() tea.Cmd {
	if m.preSelectedAgent != "" {
		for i, item := range agents {
			if a := item.(agentInfo); a.name == m.preSelectedAgent {
				m.agentList.Select(i)
				m.state = stateChatting
				return m.openNewTab(a)
			}
		}
	}
	return textinput.Blink
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			m.killAll()
			return m, tea.Quit

		case "esc", "q":
			if m.state != stateChatting {
				if len(m.tabs) == 0 {
					return m, tea.Quit
				}
				m.state = stateChatting
				return m, nil
			}

		case "enter":
			switch m.state {
			case stateSelectingAgent:
				if item := m.agentList.SelectedItem(); item != nil {
					m.state = stateChatting
					return m, m.openNewTab(item.(agentInfo))
				}
			case stateSelectingModel:
				if item := m.modelList.SelectedItem(); item != nil && len(m.tabs) > 0 {
					t := m.tabs[m.activeTab]
					modelId := item.(modelInfo).id
					m.state = stateChatting
					return m, setModelCmd(t.session, modelId)
				}
			default:
				if len(m.tabs) > 0 {
					t := m.tabs[m.activeTab]
					if content := t.input.Value(); content != "" {
						if !t.ready {
							return m, m.showToast("agent not ready yet — please wait a moment")
						}
						t.input.SetValue("")
						t.messages = append(t.messages, chatMessage{kind: kindChat, sender: "User", content: content})
						m.updateViewport(m.activeTab)
						return m, sendMessageCmd(t.session, content)
					}
				}
			}

		case "ctrl+l":
			m.showLogs = !m.showLogs
			m.updateLayout()

		case "ctrl+t":
			// Toggle between in_progress and extended think modes.
			if len(m.tabs) > 0 {
				t := m.tabs[m.activeTab]
				if t.thinkMode == thinkModeProgress {
					t.thinkMode = thinkModeExtended
				} else {
					t.thinkMode = thinkModeProgress
				}
				m.updateViewport(m.activeTab)
			}

		case "ctrl+m":
			if m.state == stateChatting && len(m.tabs) > 0 {
				m.state = stateSelectingModel
			}

		case "ctrl+a":
			m.state = stateSelectingAgent

		case "ctrl+w":
			if len(m.tabs) > 0 {
				m.closeTab(m.activeTab)
				if len(m.tabs) == 0 {
					m.state = stateSelectingAgent
				} else if m.activeTab >= len(m.tabs) {
					m.activeTab = len(m.tabs) - 1
				}
				m.updateLayout()
			}

		case "tab":
			if m.state == stateChatting && len(m.tabs) > 1 {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				m.updateLayout()
			}

		case "shift+tab":
			if m.state == stateChatting && len(m.tabs) > 1 {
				m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
				m.updateLayout()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case sessionStartedMsg:
		if _, t := m.tabForSession(msg.session); t != nil && msg.err != nil {
			t.messages = append(t.messages, chatMessage{
				kind:    kindChat,
				sender:  "System",
				content: "Failed to start agent: " + msg.err.Error(),
			})
			m.updateViewport(m.activeTab)
			cmds = append(cmds, m.showToast(msg.err.Error()))
		}

	case sessionEventMsg:
		idx, t := m.tabForSession(msg.session)
		if t == nil || msg.event.Type == EventClosed {
			return m, nil
		}

		switch msg.event.Type {

		case EventReady:
			t.ready = true
			t.messages = append(t.messages, chatMessage{
				kind:    kindChat,
				sender:  "System",
				content: "Connected — start chatting!",
			})
			m.updateViewport(idx)

		case EventMessage:
			// A real message ends the thinking phase.
			t.isThinking = false
			agentName := t.session.Info.name
			if len(t.messages) > 0 && t.messages[len(t.messages)-1].sender == agentName &&
				t.messages[len(t.messages)-1].kind == kindChat {
				t.messages[len(t.messages)-1].content += msg.event.Content
			} else {
				t.messages = append(t.messages, chatMessage{
					kind:    kindChat,
					sender:  agentName,
					content: msg.event.Content,
				})
			}
			m.updateViewport(idx)

		case EventThinking:
			t.isThinking = true
			// Accumulate into the last thought entry if one is already open.
			thinkTag := t.session.Info.name + ":thought"
			if len(t.messages) > 0 &&
				t.messages[len(t.messages)-1].kind == kindThought &&
				t.messages[len(t.messages)-1].sender == thinkTag {
				t.messages[len(t.messages)-1].content += msg.event.Content
			} else {
				t.messages = append(t.messages, chatMessage{
					kind:    kindThought,
					sender:  thinkTag,
					content: msg.event.Content,
				})
			}
			m.updateViewport(idx)

		case EventToolCall:
			toolId := msg.event.ToolCallId
			title := msg.event.Content
			status := msg.event.Status
			// Track as active if not yet terminal.
			if status != "completed" && status != "failed" {
				t.activeTools[toolId] = title
			}
			t.messages = append(t.messages, chatMessage{
				kind:       kindTool,
				sender:     "tool:" + toolId,
				content:    title,
				toolCallId: toolId,
				toolStatus: status,
			})
			m.updateViewport(idx)

		case EventToolUpdate:
			toolId := msg.event.ToolCallId
			status := msg.event.Status
			// Remove from active set once terminal.
			if status == "completed" || status == "failed" {
				delete(t.activeTools, toolId)
			} else if msg.event.Content != "" {
				t.activeTools[toolId] = msg.event.Content
			}
			// Patch the matching tool card in message history.
			for i := len(t.messages) - 1; i >= 0; i-- {
				if t.messages[i].toolCallId == toolId {
					t.messages[i].toolStatus = status
					if msg.event.Content != "" {
						t.messages[i].content = msg.event.Content
					}
					break
				}
			}
			m.updateViewport(idx)

		case EventLog:
			t.logs = append(t.logs, msg.event.Content)
			m.refreshLogPort(idx)

		case EventError:
			cmds = append(cmds, m.showToast(msg.event.Content))
		}

		cmds = append(cmds, listenForEvents(msg.session))

	case sessionSendErrMsg:
		cmds = append(cmds, m.showToast(msg.err.Error()))

	case clearToastMsg:
		m.errorToast = ""
	}

	// Delegate input to the focused sub-component.
	switch m.state {
	case stateSelectingAgent:
		var cmd tea.Cmd
		m.agentList, cmd = m.agentList.Update(msg)
		cmds = append(cmds, cmd)
	case stateSelectingModel:
		var cmd tea.Cmd
		m.modelList, cmd = m.modelList.Update(msg)
		cmds = append(cmds, cmd)
	case stateChatting:
		if len(m.tabs) > 0 {
			t := m.tabs[m.activeTab]
			var iCmd, vpCmd tea.Cmd
			t.input, iCmd = t.input.Update(msg)
			t.viewport, vpCmd = t.viewport.Update(msg)
			cmds = append(cmds, iCmd, vpCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// ── Layout ────────────────────────────────────────────────────────────────────

func (m *model) updateLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}

	// Modal list sizing.
	modalInnerW := m.width - 16
	if modalInnerW < 28 {
		modalInnerW = 28
	}
	if modalInnerW > 66 {
		modalInnerW = 66
	}
	listH := m.height - 16
	if listH < 4 {
		listH = 4
	}
	if listH > 20 {
		listH = 20
	}
	m.agentList.SetSize(modalInnerW, listH)
	m.modelList.SetSize(modalInnerW, listH)

	// Chat pane sizing.
	chatW := m.width - 2
	if m.showLogs {
		chatW = (m.width * 2 / 3) - 2
	}

	// Reserve rows: tab bar(1) + top border(1) + frame borders(2) +
	// input(1) + status line(1) + toast(1) + help(1) = 9
	vpH := m.height - 9
	if vpH < 2 {
		vpH = 2
	}
	logPortW := m.width - chatW - 6
	if logPortW < 10 {
		logPortW = 10
	}

	for i, t := range m.tabs {
		t.viewport.Width = chatW - 6
		t.viewport.Height = vpH
		t.logPort.Width = logPortW
		t.logPort.Height = m.height - 6
		t.input.Width = chatW - 8
		m.updateViewport(i)
		m.refreshLogPort(i)
	}
}

// ── Message rendering ─────────────────────────────────────────────────────────

func naturalWidth(s string) int {
	w := 0
	for _, line := range strings.Split(s, "\n") {
		if lw := len(line); lw > w {
			w = lw
		}
	}
	return w
}

// renderThoughtBubble renders an agent reasoning chunk (extended mode only).
func renderThoughtBubble(content string, vpWidth int) string {
	maxW := int(float64(vpWidth) * 0.85)
	if maxW < 20 {
		maxW = 20
	}
	label := thoughtLabelStyle.Render("  💭  thinking")
	bubble := thoughtBubbleStyle.Width(maxW).Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, label, bubble)
}

// renderToolCard renders a tool-call card with a colour-coded status indicator.
func renderToolCard(content, status string, vpWidth int) string {
	type spec struct {
		icon  string
		color lipgloss.Color
		bg    lipgloss.Color
	}
	var s spec
	switch status {
	case "pending":
		s = spec{"○", lipgloss.Color("#888822"), lipgloss.Color("#141408")}
	case "in_progress":
		s = spec{"⟳", lipgloss.Color("#00ADD8"), lipgloss.Color("#001828")}
	case "completed":
		s = spec{"✓", lipgloss.Color("#00AA44"), lipgloss.Color("#001A0A")}
	case "failed":
		s = spec{"✗", lipgloss.Color("#FF4444"), lipgloss.Color("#1A0000")}
	default:
		s = spec{"·", lipgloss.Color("#555566"), lipgloss.Color("#111118")}
	}

	maxW := vpWidth / 2
	if maxW < 28 {
		maxW = 28
	}

	icon := lipgloss.NewStyle().Foreground(s.color).Bold(true).Render(s.icon)
	title := toolLabelStyle.Render("  🔧  " + content)
	statusLabel := lipgloss.NewStyle().Foreground(s.color).Render("  " + status)

	return toolCardBaseStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.color).
		Background(s.bg).
		Width(maxW).
		Render(icon + title + statusLabel)
}

// renderMessage returns the styled string for one chatMessage, or "" if the
// message kind should not be visible in the current mode.
func (m *model) renderMessage(msg chatMessage, vpWidth int, extended bool) string {
	switch msg.kind {
	case kindThought:
		if !extended {
			return ""
		}
		return renderThoughtBubble(msg.content, vpWidth)

	case kindTool:
		if !extended {
			return ""
		}
		return renderToolCard(msg.content, msg.toolStatus, vpWidth)

	default: // kindChat
		maxBubble := int(float64(vpWidth) * 0.75)
		if maxBubble < 20 {
			maxBubble = 20
		}
		switch msg.sender {
		case "User":
			w := naturalWidth(msg.content) + 2
			if w > maxBubble {
				w = maxBubble
			}
			if w < 6 {
				w = 6
			}
			bubble := userBubbleStyle.Width(w).Render(msg.content)
			label := userLabelStyle.Render("You")
			block := lipgloss.JoinVertical(lipgloss.Right, bubble, label)
			return lipgloss.NewStyle().Width(vpWidth).Align(lipgloss.Right).Render(block)

		case "System":
			line := systemMsgStyle.Render("─── " + msg.content + " ───")
			return lipgloss.NewStyle().Width(vpWidth).Align(lipgloss.Center).Render(line)

		default: // agent
			label := agentLabelStyle.Render(msg.sender)
			bubble := agentBubbleStyle.Width(maxBubble).Render(msg.content)
			return lipgloss.JoinVertical(lipgloss.Left, label, bubble)
		}
	}
}

func (m *model) updateViewport(tabIndex int) {
	t := m.tabs[tabIndex]
	if t.viewport.Width == 0 {
		return
	}
	extended := t.thinkMode == thinkModeExtended
	var parts []string
	for _, msg := range t.messages {
		if rendered := m.renderMessage(msg, t.viewport.Width, extended); rendered != "" {
			parts = append(parts, rendered)
		}
	}
	t.viewport.SetContent(strings.Join(parts, "\n\n"))
	t.viewport.GotoBottom()
}

// ── Log panel ─────────────────────────────────────────────────────────────────

func hardWrapStr(s string, width int) []string {
	if width <= 0 || s == "" {
		return []string{s}
	}
	var lines []string
	runes := []rune(s)
	for len(runes) > width {
		lines = append(lines, string(runes[:width]))
		runes = runes[width:]
	}
	if len(runes) > 0 {
		lines = append(lines, string(runes))
	}
	return lines
}

func formatLogEntry(entry string, width int) string {
	type spec struct {
		strip  int
		tag    string
		tStyle lipgloss.Style
	}
	var s spec
	switch {
	case strings.HasPrefix(entry, "SEND: "):
		s = spec{6, "→ ", logSendStyle}
	case strings.HasPrefix(entry, "RECV: "):
		s = spec{6, "← ", logRecvStyle}
	case strings.HasPrefix(entry, "STDERR: "):
		s = spec{8, "! ", logStderrStyle}
	default:
		s = spec{0, "  ", logBodyStyle}
	}

	body := entry[s.strip:]
	tagW := 2
	bodyW := width - tagW
	if bodyW < 1 {
		bodyW = 1
	}

	lines := hardWrapStr(body, bodyW)
	var sb strings.Builder
	for i, line := range lines {
		if i == 0 {
			sb.WriteString(s.tStyle.Render(s.tag))
		} else {
			sb.WriteString("  ")
		}
		sb.WriteString(logBodyStyle.Render(line))
		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m *model) refreshLogPort(tabIndex int) {
	t := m.tabs[tabIndex]
	innerW := t.logPort.Width - 2
	if innerW < 4 {
		return
	}
	var sb strings.Builder
	for i, entry := range t.logs {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(formatLogEntry(entry, innerW))
	}
	t.logPort.SetContent(sb.String())
	t.logPort.GotoBottom()
}

func (m *model) renderLogPanel(t *tab) string {
	w := t.logPort.Width
	totalW := w + 4

	header := logHeaderStyle.Width(totalW - 2).Render(
		fmt.Sprintf(" Protocol Log  %d entries", len(t.logs)),
	)

	scrollHint := ""
	if len(t.logs) > t.logPort.Height {
		scrollHint = "  ↑/↓"
	}
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2A2A4A")).
		Render(strings.Repeat("─", totalW-2-len(scrollHint)) + scrollHint)

	inner := lipgloss.JoinVertical(lipgloss.Left, header, t.logPort.View(), footer)

	return logPanelStyle.Width(totalW).Height(t.logPort.Height + 4).Render(inner)
}

// ── Modal overlay ─────────────────────────────────────────────────────────────

func (m *model) renderModalOverlay(title, listView string) string {
	if m.width == 0 || m.height == 0 {
		return listView
	}
	heading := lipgloss.JoinVertical(lipgloss.Left,
		modalTitleStyle.Render("  ✦  "+title),
		modalHintStyle.Render("  ↑/↓ navigate  Enter select  Esc dismiss"),
		"",
	)
	modal := modalBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, heading, listView),
	)
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceChars("░"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#1A1225")),
	)
}

// ── Welcome screen ────────────────────────────────────────────────────────────

func (m *model) renderWelcome() string {
	if m.width == 0 {
		return ""
	}
	content := lipgloss.JoinVertical(lipgloss.Center,
		welcomeTitleStyle.Render("ACP Agent Client"),
		"",
		welcomeSubStyle.Render("Ctrl+A  open agent selector"),
		welcomeSubStyle.Render("Ctrl+C  quit"),
	)
	return lipgloss.NewStyle().
		Width(m.width).Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m *model) View() string {
	switch m.state {
	case stateSelectingAgent:
		return m.renderModalOverlay("Select ACP Agent", m.agentList.View())
	case stateSelectingModel:
		return m.renderModalOverlay("Select Model", m.modelList.View())
	}

	if len(m.tabs) == 0 {
		return m.renderWelcome()
	}

	active := m.tabs[m.activeTab]

	// ── Tab bar ──
	var tabLabels []string
	for i, t := range m.tabs {
		style := tabStyle
		if i == m.activeTab {
			style = activeTabStyle
		}
		label := t.session.Info.name
		if !t.ready {
			label += " ⟳"
		} else if t.thinkMode == thinkModeExtended {
			label += " 💭"
		}
		tabLabels = append(tabLabels, style.Render(label))
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabLabels...)

	// ── Custom top border with embedded model name ──
	modelDisplay := active.session.ModelId
	if modelDisplay == "" {
		modelDisplay = "default"
	}
	innerW := active.viewport.Width + 4
	modelLabel := fmt.Sprintf(" model: %s ", modelDisplay)
	dashes := innerW - len(modelLabel)
	if dashes < 0 {
		dashes = 0
	}
	topBorder := lipgloss.NewStyle().Foreground(lipgloss.Color("#3D2060")).
		Render("╭" + strings.Repeat("─", dashes) + modelLabel + "╮")

	// ── Chat content ──
	chatContent := chatFrameStyle.
		BorderTop(false).
		Width(active.viewport.Width + 2).
		Height(active.viewport.Height + 1).
		Render(active.viewport.View())

	// ── Input row ──
	inputLine := lipgloss.NewStyle().Padding(0, 1).
		Render(inputPromptStyle.Render("> ") + active.input.View())

	// ── Status line (always 1 row to avoid layout shift) ──
	// In extended mode: show a mode badge.
	// In progress mode: show live activity or blank.
	var statusLine string
	if active.thinkMode == thinkModeExtended {
		statusLine = extendedModeStyle.Width(m.width).
			Render("  💭  extended thinking  —  Ctrl+T to collapse")
	} else {
		activity := active.activityStatus()
		if activity != "" {
			statusLine = activityBarStyle.Width(m.width).Render(activity)
		} else {
			statusLine = lipgloss.NewStyle().Width(m.width).Render(" ")
		}
	}

	// ── Toast bar (always 1 row) ──
	toastLine := lipgloss.NewStyle().Width(m.width).Render(" ")
	if m.errorToast != "" {
		toastLine = toastStyle.Width(m.width).Render("  ⚠  " + m.errorToast)
	}

	// ── Help bar ──
	helpLine := helpStyle.Width(m.width).Render(
		"Tab: next  •  ^W: close  •  ^A: agent  •  ^M: model  •  ^T: thinking  •  ^L: logs  •  ^C: quit",
	)

	chatCol := lipgloss.JoinVertical(lipgloss.Left,
		tabBar, topBorder, chatContent, inputLine,
	)

	mainArea := chatCol
	if m.showLogs {
		mainArea = lipgloss.JoinHorizontal(lipgloss.Top, chatCol, m.renderLogPanel(active))
	}

	return lipgloss.JoinVertical(lipgloss.Left, mainArea, statusLine, toastLine, helpLine)
}

// ── Entry points ──────────────────────────────────────────────────────────────

func startTui() {
	startTuiWithAgent("")
}

func startTuiWithAgent(agentName string) {
	al := list.New(agents, list.NewDefaultDelegate(), 0, 0)
	al.SetShowTitle(false)
	al.SetShowStatusBar(false)
	al.SetFilteringEnabled(true)

	ml := list.New(defaultModels, list.NewDefaultDelegate(), 0, 0)
	ml.SetShowTitle(false)
	ml.SetShowStatusBar(false)
	ml.SetFilteringEnabled(true)

	m := &model{
		state:            stateSelectingAgent,
		agentList:        al,
		modelList:        ml,
		preSelectedAgent: agentName,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		panic(err)
	}
}
