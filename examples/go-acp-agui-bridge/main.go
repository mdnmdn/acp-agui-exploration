package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	port      string
	agentCmd  string
	agentArgs []string
	verbose   bool
	cwd       string
)

var rootCmd = &cobra.Command{
	Use:   "go-acp-agui-bridge",
	Short: "Bridge ACP agents to AG-UI protocol over HTTP",
	Long: `Run an ACP-compatible agent and expose it via AG-UI protocol on HTTP.
The server provides a web interface and AG-UI endpoint for interacting with the agent.`,
	Run: func(cmd *cobra.Command, args []string) {
		if agentCmd == "" {
			fmt.Println("Error: --agent flag is required")
			os.Exit(1)
		}

		if cwd == "" {
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Printf("Starting ACP agent: %s %v", agentCmd, agentArgs)
		log.Printf("Server running on :%s", port)

		mux := http.NewServeMux()
		mux.Handle("/agent", handleAgentRun(agentCmd, agentArgs, cwd))
		mux.Handle("/", handleStatic())

		log.Fatal(http.ListenAndServe(":"+port, mux))
	},
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "3000", "Server port")
	rootCmd.Flags().StringVarP(&agentCmd, "agent", "a", "", "Agent command to run (required)")
	rootCmd.Flags().StringSliceVarP(&agentArgs, "agent-args", "", []string{"--acp"}, "Agent arguments")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().StringVar(&cwd, "cwd", "", "Working directory for the agent")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type acpSession struct {
	client    *acp.ClientSideConnection
	cmd       *exec.Cmd
	sessionId string
	events    chan string
}

func newACPSession(ctx context.Context, agentCmd string, agentArgs []string, cwd string) (*acpSession, error) {
	cmd := exec.CommandContext(ctx, agentCmd, agentArgs...)
	cmd.Dir = cwd

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if verbose {
		go func() {
			buf := make([]byte, 1024)
			for {
				n, rerr := stderr.Read(buf)
				if n > 0 {
					log.Printf("[AGENT STDERR] %s", string(buf[:n]))
				}
				if rerr != nil {
					return
				}
			}
		}()
	}

	client := &acpClient{events: make(chan string)}
	conn := acp.NewClientSideConnection(client, stdin, stdout)

	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	initResp, err := conn.Initialize(ctx2, acp.InitializeRequest{
		ClientCapabilities: acp.ClientCapabilities{},
		ClientInfo:         &acp.Implementation{Name: "acp-agui-bridge", Version: "0.1.0"},
		ProtocolVersion:    acp.ProtocolVersionNumber,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize failed: %w", err)
	}

	if len(initResp.AuthMethods) > 0 {
		if _, err = conn.Authenticate(ctx2, acp.AuthenticateRequest{
			MethodId: initResp.AuthMethods[0].Id,
		}); err != nil {
			return nil, fmt.Errorf("authenticate failed: %w", err)
		}
	}

	resp, err := conn.NewSession(ctx2, acp.NewSessionRequest{
		Cwd:        cwd,
		McpServers: []acp.McpServer{},
	})
	if err != nil {
		return nil, fmt.Errorf("new session failed: %w", err)
	}

	return &acpSession{
		client:    conn,
		cmd:       cmd,
		sessionId: string(resp.SessionId),
		events:    client.events,
	}, nil
}

func (s *acpSession) close() {
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}
}

type acpClient struct {
	events chan string
}

func (c *acpClient) ReadTextFile(_ context.Context, _ acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	return acp.ReadTextFileResponse{}, nil
}
func (c *acpClient) WriteTextFile(_ context.Context, _ acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	return acp.WriteTextFileResponse{}, nil
}
func (c *acpClient) RequestPermission(_ context.Context, _ acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	return acp.RequestPermissionResponse{}, nil
}

func (c *acpClient) SessionUpdate(_ context.Context, params acp.SessionNotification) error {
	u := params.Update
	switch {
	case u.AgentMessageChunk != nil && u.AgentMessageChunk.Content.Text != nil:
		c.events <- fmt.Sprintf(`{"type":"TEXT_MESSAGE_CONTENT","delta":"%s"}`, escapeJSON(u.AgentMessageChunk.Content.Text.Text))
	case u.AgentThoughtChunk != nil && u.AgentThoughtChunk.Content.Text != nil:
		c.events <- fmt.Sprintf(`{"type":"REASONING_MESSAGE_CONTENT","delta":"%s"}`, escapeJSON(u.AgentThoughtChunk.Content.Text.Text))
	case u.ToolCall != nil:
		status := string(u.ToolCall.Status)
		c.events <- fmt.Sprintf(`{"type":"TOOL_CALL_START","toolCallId":"%s","toolCallName":"%s","status":"%s"}`, u.ToolCall.ToolCallId, u.ToolCall.Title, status)
	case u.ToolCallUpdate != nil:
		status := ""
		if u.ToolCallUpdate.Status != nil {
			status = string(*u.ToolCallUpdate.Status)
		}
		c.events <- fmt.Sprintf(`{"type":"TOOL_CALL_ARGS","toolCallId":"%s","delta":"%s"}`, u.ToolCallUpdate.ToolCallId, status)
	}
	return nil
}

func (c *acpClient) CreateTerminal(_ context.Context, _ acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	return acp.CreateTerminalResponse{}, nil
}
func (c *acpClient) KillTerminalCommand(_ context.Context, _ acp.KillTerminalCommandRequest) (acp.KillTerminalCommandResponse, error) {
	return acp.KillTerminalCommandResponse{}, nil
}
func (c *acpClient) TerminalOutput(_ context.Context, _ acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{}, nil
}
func (c *acpClient) ReleaseTerminal(_ context.Context, _ acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, nil
}
func (c *acpClient) WaitForTerminalExit(_ context.Context, _ acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{}, nil
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

func handleAgentRun(agentCmd string, agentArgs []string, cwd string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		accept := r.Header.Get("Accept")
		if !strings.Contains(accept, "text/event-stream") && !strings.Contains(accept, "*/*") {
			http.Error(w, "Accept: text/event-stream required", http.StatusNotAcceptable)
			return
		}

		var body struct {
			ThreadID string `json:"threadId"`
			RunID    string `json:"runId"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}

		if err := decodeJSON(r.Body, &body); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if body.ThreadID == "" {
			body.ThreadID = uuid.New().String()
		}
		if body.RunID == "" {
			body.RunID = uuid.New().String()
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}

		sendEvent := func(data string) {
			w.Write([]byte("data: " + data + "\n\n"))
			flusher.Flush()
		}

		sendEvent(fmt.Sprintf(`{"type":"RUN_STARTED","threadId":"%s","runId":"%s"}`, body.ThreadID, body.RunID))

		var userMsg string
		for _, msg := range body.Messages {
			if msg.Role == "user" {
				userMsg = msg.Content
				break
			}
		}

		if userMsg == "" {
			sendEvent(`{"type":"RUN_ERROR","message":"No user message provided"}`)
			return
		}

		ctx := r.Context()
		session, err := newACPSession(ctx, agentCmd, agentArgs, cwd)
		if err != nil {
			sendEvent(fmt.Sprintf(`{"type":"RUN_ERROR","message":"%s"}`, escapeJSON(err.Error())))
			return
		}
		defer session.close()

		messageID := uuid.New().String()
		sendEvent(fmt.Sprintf(`{"type":"TEXT_MESSAGE_START","messageId":"%s","role":"assistant"}`, messageID))

		_, err = session.client.Prompt(ctx, acp.PromptRequest{
			SessionId: acp.SessionId(session.sessionId),
			Prompt:    []acp.ContentBlock{acp.TextBlock(userMsg)},
		})
		if err != nil {
			sendEvent(fmt.Sprintf(`{"type":"RUN_ERROR","message":"%s"}`, escapeJSON(err.Error())))
			return
		}

		done := ctx.Done()
		for {
			select {
			case event := <-session.events:
				sendEvent(event)
			case <-done:
				goto finish
			}
		}

	finish:
		sendEvent(fmt.Sprintf(`{"type":"TEXT_MESSAGE_END","messageId":"%s"}`, messageID))
		sendEvent(fmt.Sprintf(`{"type":"RUN_FINISHED","threadId":"%s","runId":"%s"}`, body.ThreadID, body.RunID))
	})
}

func handleStatic() http.Handler {
	fs := http.FileServer(http.Dir("./static"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
	})
}

func decodeJSON(data io.Reader, v interface{}) error {
	decoder := json.NewDecoder(data)
	decoder.UseNumber()
	return decoder.Decode(v)
}
