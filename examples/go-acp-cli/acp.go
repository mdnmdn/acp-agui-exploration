package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	acp "github.com/coder/acp-go-sdk"
)

// ── Event stream ──────────────────────────────────────────────────────────────

// AgentEventType categorizes events published by an AgentSession.
type AgentEventType int

const (
	EventMessage    AgentEventType = iota // agent sent a text chunk
	EventThinking                         // agent reasoning / thought chunk
	EventToolCall                         // tool invocation started
	EventToolUpdate                       // tool invocation status changed
	EventLog                              // raw protocol byte log
	EventError                            // error inside the session
	EventReady                            // ACP handshake complete
	EventClosed                           // session shut down
)

// AgentEvent is published on the session's Events() channel.
// ToolCallId and Status are only populated for EventToolCall / EventToolUpdate.
type AgentEvent struct {
	Type       AgentEventType
	Content    string // text chunk, tool title, or empty
	ToolCallId string // tool call identifier
	Status     string // "pending" | "in_progress" | "completed" | "failed"
}

// ── AgentSession ──────────────────────────────────────────────────────────────

// AgentSession manages one ACP subprocess connection.
// It is fully decoupled from any UI framework — all output flows through Events().
//
// Frontier contract (what callers outside this file may use):
//   - NewAgentSession / Start / Send / SetModel / Close
//   - Events() for reading output
//   - SessionId / ModelId string fields (read after Start returns)
type AgentSession struct {
	Info      agentInfo
	SessionId string
	ModelId   string

	events    chan AgentEvent
	done      chan struct{}
	closeOnce sync.Once
	acpClient *acp.ClientSideConnection
	agentCmd  *exec.Cmd
	mu        sync.Mutex
}

// NewAgentSession allocates a session. Call Start to launch the subprocess.
func NewAgentSession(info agentInfo) *AgentSession {
	return &AgentSession{
		Info:   info,
		events: make(chan AgentEvent, 256),
		done:   make(chan struct{}),
	}
}

// Events returns a read-only stream of AgentEvents.
func (s *AgentSession) Events() <-chan AgentEvent {
	return s.events
}

// emit writes an event without blocking; silently drops it if the session is closing.
func (s *AgentSession) emit(e AgentEvent) {
	select {
	case s.events <- e:
	case <-s.done:
	}
}

// Start launches the agent subprocess, performs the ACP handshake, and opens a session.
// Blocks until the session is established or an error occurs.
// On success it emits EventReady; on failure it returns an error.
func (s *AgentSession) Start(ctx context.Context, cwd string) error {
	execCmd := s.Info.command
	if strings.HasPrefix(execCmd, "./") {
		if base := strings.TrimPrefix(execCmd, "./"); isInPath(base) {
			execCmd = base
		}
	}

	args := s.Info.args
	if execCmd == "opencode" && len(args) == 0 {
		args = []string{"acp"}
	}

	cmd := exec.CommandContext(ctx, execCmd, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.agentCmd = cmd

	go func() {
		buf := make([]byte, 1024)
		for {
			n, rerr := stderr.Read(buf)
			if n > 0 {
				s.emit(AgentEvent{Type: EventLog, Content: "STDERR: " + string(buf[:n])})
			}
			if rerr != nil {
				return
			}
		}
	}()

	client := &sessionClient{session: s}
	conn := acp.NewClientSideConnection(
		client,
		&loggingWriter{Writer: stdin, session: s},
		&loggingReader{Reader: stdout, session: s},
	)
	s.acpClient = conn

	initResp, err := conn.Initialize(ctx, acp.InitializeRequest{
		ClientCapabilities: acp.ClientCapabilities{},
		ClientInfo:         &acp.Implementation{Name: "acp-client", Version: "0.1.0"},
		ProtocolVersion:    acp.ProtocolVersionNumber,
	})
	if err != nil {
		return err
	}

	if len(initResp.AuthMethods) > 0 {
		if _, err = conn.Authenticate(ctx, acp.AuthenticateRequest{
			MethodId: initResp.AuthMethods[0].Id,
		}); err != nil {
			return err
		}
	}

	resp, err := conn.NewSession(ctx, acp.NewSessionRequest{
		Cwd:        cwd,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return err
	}
	if resp.SessionId == "" {
		return fmt.Errorf("agent returned empty session ID")
	}
	s.SessionId = string(resp.SessionId)

	s.emit(AgentEvent{Type: EventReady})
	return nil
}

// Send dispatches a user prompt to the agent.
// The response arrives asynchronously as EventMessage events on Events().
func (s *AgentSession) Send(ctx context.Context, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.acpClient == nil || s.SessionId == "" {
		return fmt.Errorf("session not ready")
	}
	_, err := s.acpClient.Prompt(ctx, acp.PromptRequest{
		SessionId: acp.SessionId(s.SessionId),
		Prompt:    []acp.ContentBlock{acp.TextBlock(content)},
	})
	return err
}

// SetModel changes the active model for this session.
func (s *AgentSession) SetModel(ctx context.Context, modelId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.acpClient == nil || s.SessionId == "" {
		return fmt.Errorf("session not ready")
	}
	_, err := s.acpClient.SetSessionModel(ctx, acp.SetSessionModelRequest{
		SessionId: acp.SessionId(s.SessionId),
		ModelId:   acp.ModelId(modelId),
	})
	if err == nil {
		s.ModelId = modelId
	}
	return err
}

// Close terminates the subprocess and unblocks any goroutine blocked on Events().
func (s *AgentSession) Close() {
	s.closeOnce.Do(func() {
		close(s.done)
		if s.agentCmd != nil && s.agentCmd.Process != nil {
			_ = s.agentCmd.Process.Kill()
		}
		select {
		case s.events <- AgentEvent{Type: EventClosed}:
		default:
		}
	})
}

// ── acp.Client implementation ─────────────────────────────────────────────────

type sessionClient struct{ session *AgentSession }

func (c *sessionClient) ReadTextFile(_ context.Context, _ acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	return acp.ReadTextFileResponse{}, nil
}
func (c *sessionClient) WriteTextFile(_ context.Context, _ acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	return acp.WriteTextFileResponse{}, nil
}
func (c *sessionClient) RequestPermission(_ context.Context, _ acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	return acp.RequestPermissionResponse{}, nil
}

func (c *sessionClient) SessionUpdate(_ context.Context, params acp.SessionNotification) error {
	u := params.Update
	switch {
	case u.AgentMessageChunk != nil && u.AgentMessageChunk.Content.Text != nil:
		c.session.emit(AgentEvent{
			Type:    EventMessage,
			Content: u.AgentMessageChunk.Content.Text.Text,
		})

	case u.AgentThoughtChunk != nil && u.AgentThoughtChunk.Content.Text != nil:
		c.session.emit(AgentEvent{
			Type:    EventThinking,
			Content: u.AgentThoughtChunk.Content.Text.Text,
		})

	case u.ToolCall != nil:
		c.session.emit(AgentEvent{
			Type:       EventToolCall,
			Content:    u.ToolCall.Title,
			ToolCallId: string(u.ToolCall.ToolCallId),
			Status:     string(u.ToolCall.Status),
		})

	case u.ToolCallUpdate != nil:
		status := ""
		if u.ToolCallUpdate.Status != nil {
			status = string(*u.ToolCallUpdate.Status)
		}
		title := ""
		if u.ToolCallUpdate.Title != nil {
			title = *u.ToolCallUpdate.Title
		}
		c.session.emit(AgentEvent{
			Type:       EventToolUpdate,
			Content:    title,
			ToolCallId: string(u.ToolCallUpdate.ToolCallId),
			Status:     status,
		})
	}
	return nil
}

func (c *sessionClient) CreateTerminal(_ context.Context, _ acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	return acp.CreateTerminalResponse{}, nil
}
func (c *sessionClient) KillTerminalCommand(_ context.Context, _ acp.KillTerminalCommandRequest) (acp.KillTerminalCommandResponse, error) {
	return acp.KillTerminalCommandResponse{}, nil
}
func (c *sessionClient) TerminalOutput(_ context.Context, _ acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{}, nil
}
func (c *sessionClient) ReleaseTerminal(_ context.Context, _ acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, nil
}
func (c *sessionClient) WaitForTerminalExit(_ context.Context, _ acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{}, nil
}

// ── Protocol loggers ──────────────────────────────────────────────────────────

type loggingReader struct {
	io.Reader
	session *AgentSession
}

func (l *loggingReader) Read(p []byte) (int, error) {
	n, err := l.Reader.Read(p)
	if n > 0 {
		l.session.emit(AgentEvent{Type: EventLog, Content: "RECV: " + string(p[:n])})
	}
	return n, err
}

type loggingWriter struct {
	io.Writer
	session *AgentSession
}

func (l *loggingWriter) Write(p []byte) (int, error) {
	n, err := l.Writer.Write(p)
	if n > 0 {
		l.session.emit(AgentEvent{Type: EventLog, Content: "SEND: " + string(p[:n])})
	}
	return n, err
}

// ── CLI (non-interactive) mode ────────────────────────────────────────────────

func runCliMode(agentName, promptText string) error {
	var selected agentInfo
	for _, item := range agents {
		if a := item.(agentInfo); a.name == agentName {
			selected = a
			break
		}
	}
	if selected.name == "" {
		return fmt.Errorf("agent not found: %s", agentName)
	}
	if !isAgentAvailable(selected) {
		return fmt.Errorf("agent %q is not installed or not in PATH", selected.name)
	}

	timeout := 30 * time.Second
	if selected.name == "opencode" {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	s := NewAgentSession(selected)
	if err := s.Start(ctx, cwd); err != nil {
		return err
	}
	if err := s.Send(ctx, promptText); err != nil {
		return err
	}

	for {
		select {
		case e := <-s.Events():
			switch e.Type {
			case EventMessage:
				fmt.Printf("Agent: %s\n", e.Content)
				s.Close()
				return nil
			case EventError:
				return fmt.Errorf("%s", e.Content)
			case EventClosed:
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ── Mock agent (built-in echo agent with simulated thinking + tools) ──────────

type mockAgent struct{ conn *acp.AgentSideConnection }

func (a *mockAgent) Authenticate(_ context.Context, _ acp.AuthenticateRequest) (acp.AuthenticateResponse, error) {
	return acp.AuthenticateResponse{}, nil
}
func (a *mockAgent) Initialize(_ context.Context, _ acp.InitializeRequest) (acp.InitializeResponse, error) {
	return acp.InitializeResponse{
		ProtocolVersion:   acp.ProtocolVersionNumber,
		AgentCapabilities: acp.AgentCapabilities{},
		AgentInfo:         &acp.Implementation{Name: "mock-agent", Version: "0.1.0"},
	}, nil
}
func (a *mockAgent) Cancel(_ context.Context, _ acp.CancelNotification) error { return nil }
func (a *mockAgent) NewSession(_ context.Context, _ acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	return acp.NewSessionResponse{SessionId: "mock-session"}, nil
}

func (a *mockAgent) Prompt(ctx context.Context, params acp.PromptRequest) (acp.PromptResponse, error) {
	text := ""
	if len(params.Prompt) > 0 && params.Prompt[0].Text != nil {
		text = params.Prompt[0].Text.Text
	}

	sid := params.SessionId

	// 1. Simulate a reasoning / thought chunk.
	_ = a.conn.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: sid,
		Update: acp.SessionUpdate{
			AgentThoughtChunk: &acp.SessionUpdateAgentThoughtChunk{
				Content: acp.TextBlock(
					"The user said: \"" + text + "\". " +
						"This is a mock agent, so the right answer is to echo it back. " +
						"No further reasoning required.",
				),
			},
		},
	})

	// 2. Simulate a tool call (in_progress → completed).
	toolId := acp.ToolCallId("mock-echo-1")
	_ = a.conn.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: sid,
		Update: acp.SessionUpdate{
			ToolCall: &acp.SessionUpdateToolCall{
				ToolCallId: toolId,
				Title:      "echo \"" + text + "\"",
				Status:     acp.ToolCallStatusInProgress,
			},
		},
	})

	completed := acp.ToolCallStatusCompleted
	_ = a.conn.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: sid,
		Update: acp.SessionUpdate{
			ToolCallUpdate: &acp.SessionToolCallUpdate{
				ToolCallId: toolId,
				Status:     &completed,
			},
		},
	})

	// 3. Final text response.
	_ = a.conn.SessionUpdate(ctx, acp.SessionNotification{
		SessionId: sid,
		Update: acp.SessionUpdate{
			AgentMessageChunk: &acp.SessionUpdateAgentMessageChunk{
				Content: acp.TextBlock("Echo: " + text),
			},
		},
	})

	return acp.PromptResponse{}, nil
}

func (a *mockAgent) SetSessionMode(_ context.Context, _ acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	return acp.SetSessionModeResponse{}, nil
}
func (a *mockAgent) SetSessionModel(_ context.Context, _ acp.SetSessionModelRequest) (acp.SetSessionModelResponse, error) {
	return acp.SetSessionModelResponse{}, nil
}

func runMockAgent() {
	agent := &mockAgent{}
	conn := acp.NewAgentSideConnection(agent, os.Stdout, os.Stdin)
	agent.conn = conn
	<-conn.Done()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func isInPath(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func isAgentAvailable(info agentInfo) bool {
	cmd := info.command
	switch {
	case cmd == "npx" || cmd == "uvx":
		return isInPath(cmd)
	case strings.Contains(cmd, "/"):
		if _, err := os.Stat(cmd); err != nil {
			if strings.HasPrefix(cmd, "./") {
				return isInPath(strings.TrimPrefix(cmd, "./"))
			}
			return false
		}
		return true
	default:
		return isInPath(cmd)
	}
}
