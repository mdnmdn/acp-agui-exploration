package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/acp-go-sdk"
)

func runCliMode(agentName, message string) error {
	var selected agentInfo
	found := false
	for _, item := range agents {
		a := item.(agentInfo)
		if a.name == agentName {
			selected = a
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("agent not found: %s", agentName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, selected.command, selected.args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	client := &cliClient{done: make(chan struct{})}
	conn := acp.NewClientSideConnection(client, stdin, stdout)

	// Initialize ACP
	_, err = conn.Initialize(ctx, acp.InitializeRequest{
		ClientCapabilities: acp.ClientCapabilities{},
		ClientInfo: &acp.Implementation{
			Name:    "sample-acp-cli",
			Version: "0.1.0",
		},
		ProtocolVersion: acp.ProtocolVersionNumber,
	})
	if err != nil {
		return err
	}

	// Start session
	resp, err := conn.NewSession(ctx, acp.NewSessionRequest{
		Cwd:        cwd,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return err
	}

	// Send prompt
	_, err = conn.Prompt(ctx, acp.PromptRequest{
		SessionId: resp.SessionId,
		Prompt: []acp.ContentBlock{
			acp.TextBlock(message),
		},
	})
	if err != nil {
		return err
	}

	// Wait for response or timeout
	select {
	case <-client.done:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

type cliClient struct {
	done chan struct{}
}

func (c *cliClient) ReadTextFile(ctx context.Context, params acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	return acp.ReadTextFileResponse{}, nil
}
func (c *cliClient) WriteTextFile(ctx context.Context, params acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	return acp.WriteTextFileResponse{}, nil
}
func (c *cliClient) RequestPermission(ctx context.Context, params acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	return acp.RequestPermissionResponse{}, nil
}
func (c *cliClient) SessionUpdate(ctx context.Context, params acp.SessionNotification) error {
	update := params.Update
	if update.AgentMessageChunk != nil && update.AgentMessageChunk.Content.Text != nil {
		fmt.Printf("Agent: %s\n", update.AgentMessageChunk.Content.Text.Text)
		close(c.done)
	}
	return nil
}
func (c *cliClient) CreateTerminal(ctx context.Context, params acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	return acp.CreateTerminalResponse{}, nil
}
func (c *cliClient) KillTerminalCommand(ctx context.Context, params acp.KillTerminalCommandRequest) (acp.KillTerminalCommandResponse, error) {
	return acp.KillTerminalCommandResponse{}, nil
}
func (c *cliClient) TerminalOutput(ctx context.Context, params acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{}, nil
}
func (c *cliClient) ReleaseTerminal(ctx context.Context, params acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, nil
}
func (c *cliClient) WaitForTerminalExit(ctx context.Context, params acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{}, nil
}

func (m *model) startAgent(selected agentInfo) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			ctx := context.Background()

			cwd, err := os.Getwd()
			if err != nil {
				return errMsg{err}
			}

			cmd := exec.CommandContext(ctx, selected.command, selected.args...)

			stdin, err := cmd.StdinPipe()
			if err != nil {
				return errMsg{err}
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return errMsg{err}
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				return errMsg{err}
			}

			if err := cmd.Start(); err != nil {
				return errMsg{err}
			}

			// Initialize tab
			ti := textinput.New()
			ti.Placeholder = "Type a message..."
			ti.Focus()

			tabIndex := len(m.tabs)
			newTab := &tab{
				selected: selected,
				input:    ti,
				viewport: viewport.New(0, 0),
				logPort:  viewport.New(0, 0),
				agentCmd: cmd,
			}
			m.tabs = append(m.tabs, newTab)
			m.activeTab = tabIndex
			m.updateLayout()

			// Create a wrapper that logs everything
			loggingReader := &loggingReader{
				Reader:   stdout,
				log:      m.msgChan,
				tabIndex: tabIndex,
			}
			loggingWriter := &loggingWriter{
				Writer:   stdin,
				log:      m.msgChan,
				tabIndex: tabIndex,
			}

			client := &tuiClient{msgChan: m.msgChan, tabIndex: tabIndex}
			conn := acp.NewClientSideConnection(client, loggingWriter, loggingReader)
			newTab.acpClient = conn

			// Read stderr in background
			go func() {
				buf := make([]byte, 1024)
				for {
					n, err := stderr.Read(buf)
					if n > 0 {
						m.msgChan <- message{tabIndex: tabIndex, sender: "AgentStderr", content: string(buf[:n]), isLog: true}
					}
					if err != nil {
						break
					}
				}
			}()

			// Initialize ACP
			_, err = conn.Initialize(ctx, acp.InitializeRequest{
				ClientCapabilities: acp.ClientCapabilities{},
				ClientInfo: &acp.Implementation{
					Name:    "sample-acp-tui",
					Version: "0.1.0",
				},
				ProtocolVersion: acp.ProtocolVersionNumber,
			})
			if err != nil {
				return errMsg{err}
			}

			// Start session
			resp, err := conn.NewSession(ctx, acp.NewSessionRequest{
				Cwd:        cwd,
				McpServers: []acp.McpServer{},
			})
			if err != nil {
				return errMsg{err}
			}
			newTab.sessionId = resp.SessionId

			return successMsg{tabIndex: tabIndex}
		},
		m.waitForMsg(),
	)
}

func (m *model) setAgentModel(tabIndex int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		t := m.tabs[tabIndex]
		_, err := t.acpClient.SetSessionModel(ctx, acp.SetSessionModelRequest{
			SessionId: t.sessionId,
			ModelId:   t.modelId,
		})
		if err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (m *model) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		return <-m.msgChan
	}
}

func (m *model) sendACPMessage(tabIndex int, content string) tea.Cmd {
	return func() tea.Msg {
		t := m.tabs[tabIndex]
		t.acpMu.Lock()
		defer t.acpMu.Unlock()

		if t.acpClient == nil {
			return errMsg{fmt.Errorf("ACP client not initialized")}
		}

		_, err := t.acpClient.Prompt(context.Background(), acp.PromptRequest{
			SessionId: t.sessionId,
			Prompt: []acp.ContentBlock{
				acp.TextBlock(content),
			},
		})
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

// tuiClient implements acp.Client
type tuiClient struct {
	msgChan  chan message
	tabIndex int
}

func (c *tuiClient) ReadTextFile(ctx context.Context, params acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	return acp.ReadTextFileResponse{}, nil
}
func (c *tuiClient) WriteTextFile(ctx context.Context, params acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	return acp.WriteTextFileResponse{}, nil
}
func (c *tuiClient) RequestPermission(ctx context.Context, params acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	return acp.RequestPermissionResponse{}, nil
}
func (c *tuiClient) SessionUpdate(ctx context.Context, params acp.SessionNotification) error {
	update := params.Update
	if update.AgentMessageChunk != nil && update.AgentMessageChunk.Content.Text != nil {
		c.msgChan <- message{tabIndex: c.tabIndex, sender: "Agent", content: update.AgentMessageChunk.Content.Text.Text}
	}
	return nil
}
func (c *tuiClient) CreateTerminal(ctx context.Context, params acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	return acp.CreateTerminalResponse{}, nil
}
func (c *tuiClient) KillTerminalCommand(ctx context.Context, params acp.KillTerminalCommandRequest) (acp.KillTerminalCommandResponse, error) {
	return acp.KillTerminalCommandResponse{}, nil
}
func (c *tuiClient) TerminalOutput(ctx context.Context, params acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{}, nil
}
func (c *tuiClient) ReleaseTerminal(ctx context.Context, params acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, nil
}
func (c *tuiClient) WaitForTerminalExit(ctx context.Context, params acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{}, nil
}

type loggingReader struct {
	io.Reader
	log      chan message
	tabIndex int
}

func (l *loggingReader) Read(p []byte) (n int, err error) {
	n, err = l.Reader.Read(p)
	if n > 0 {
		l.log <- message{tabIndex: l.tabIndex, content: fmt.Sprintf("RECV: %s", string(p[:n])), isLog: true}
	}
	return
}

type loggingWriter struct {
	io.Writer
	log      chan message
	tabIndex int
}

func (l *loggingWriter) Write(p []byte) (n int, err error) {
	n, err = l.Writer.Write(p)
	if n > 0 {
		l.log <- message{tabIndex: l.tabIndex, content: fmt.Sprintf("SEND: %s", string(p[:n])), isLog: true}
	}
	return
}

// Mock Agent Implementation
type mockAgent struct {
	conn *acp.AgentSideConnection
}

func (a *mockAgent) Authenticate(ctx context.Context, params acp.AuthenticateRequest) (acp.AuthenticateResponse, error) {
	return acp.AuthenticateResponse{}, nil
}
func (a *mockAgent) Initialize(ctx context.Context, params acp.InitializeRequest) (acp.InitializeResponse, error) {
	return acp.InitializeResponse{
		ProtocolVersion: acp.ProtocolVersionNumber,
		AgentCapabilities: acp.AgentCapabilities{},
		AgentInfo: &acp.Implementation{
			Name:    "mock-agent",
			Version: "0.1.0",
		},
	}, nil
}
func (a *mockAgent) Cancel(ctx context.Context, params acp.CancelNotification) error {
	return nil
}
func (a *mockAgent) NewSession(ctx context.Context, params acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	return acp.NewSessionResponse{SessionId: "mock-session"}, nil
}
func (a *mockAgent) Prompt(ctx context.Context, params acp.PromptRequest) (acp.PromptResponse, error) {
	// Echo back via SessionUpdate
	text := ""
	if len(params.Prompt) > 0 && params.Prompt[0].Text != nil {
		text = params.Prompt[0].Text.Text
	}
	_ = a.conn.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: params.SessionId,
		Update: acp.SessionUpdate{
			AgentMessageChunk: &acp.SessionUpdateAgentMessageChunk{
				Content: acp.TextBlock("Echo: " + text),
			},
		},
	})
	return acp.PromptResponse{}, nil
}
func (a *mockAgent) SetSessionMode(ctx context.Context, params acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	return acp.SetSessionModeResponse{}, nil
}
func (a *mockAgent) SetSessionModel(ctx context.Context, params acp.SetSessionModelRequest) (acp.SetSessionModelResponse, error) {
	return acp.SetSessionModelResponse{}, nil
}

func runMockAgent() {
	agent := &mockAgent{}
	conn := acp.NewAgentSideConnection(agent, os.Stdout, os.Stdin)
	agent.conn = conn
	<-conn.Done()
}
