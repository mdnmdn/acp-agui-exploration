package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type Request struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
	Body   json.RawMessage `json:"body,omitempty"`
}

type AgentParams struct {
	AgentID string `json:"agentId"`
}

type Message struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AgentBody struct {
	ThreadID       string                 `json:"threadId"`
	RunID          string                 `json:"runId"`
	Tools          []interface{}          `json:"tools"`
	Context        []interface{}          `json:"context"`
	ForwardedProps map[string]interface{} `json:"forwardedProps"`
	State          map[string]interface{} `json:"state"`
	Messages       []Message              `json:"messages"`
}

type Agent interface {
	Info() map[string]interface{}
	Connect(body AgentBody) error
	Run(body AgentBody, stream SSEWriter) error
}

type SSEWriter interface {
	Write(data interface{}) error
}

type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (s *sseWriter) Write(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.w.Write([]byte("data: " + string(jsonData) + "\n\n"))
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

type Protocol struct {
	agent   Agent
	storage Storage
}

func NewProtocol(agent Agent, storage Storage) *Protocol {
	return &Protocol{agent: agent, storage: storage}
}

func (p *Protocol) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch req.Method {
	case "info":
		p.handleInfo(w)
	case "agent/connect":
		p.handleAgentConnect(w, req)
	case "agent/run":
		p.handleAgentRun(w, req)
	default:
		http.Error(w, "unknown method", http.StatusBadRequest)
	}
}

func (p *Protocol) handleInfo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p.agent.Info())
}

func (p *Protocol) handleAgentConnect(w http.ResponseWriter, req Request) {
	var body AgentBody
	if err := json.Unmarshal(req.Body, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	if err := p.agent.Connect(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.(http.Flusher).Flush()
}

func (p *Protocol) handleAgentRun(w http.ResponseWriter, req Request) {
	var body AgentBody
	if err := json.Unmarshal(req.Body, &body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Load full thread history
	thread, _ := p.storage.GetThread(body.ThreadID)
	existingIDs := make(map[string]bool)
	if thread != nil {
		for _, msg := range thread.Messages {
			existingIDs[msg.ID] = true
		}
	}

	// Save only new user messages
	for _, msg := range body.Messages {
		if msg.Role == "user" && !existingIDs[msg.ID] {
			p.storage.SaveMessage(body.ThreadID, body.RunID, msg.ID, msg.Role, msg.Content)
		}
	}

	// Reload thread with new messages
	thread, _ = p.storage.GetThread(body.ThreadID)
	if thread != nil {
		body.Messages = make([]Message, len(thread.Messages))
		for i, msg := range thread.Messages {
			body.Messages[i] = Message{
				ID:      msg.ID,
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	stream := &sseWriter{w: w, flusher: w.(http.Flusher)}

	stream.Write(map[string]interface{}{
		"type":     "RUN_STARTED",
		"threadId": body.ThreadID,
		"runId":    body.RunID,
		"input": map[string]interface{}{
			"threadId":       body.ThreadID,
			"runId":          body.RunID,
			"state":          body.State,
			"messages":       body.Messages,
			"tools":          body.Tools,
			"context":        body.Context,
			"forwardedProps": body.ForwardedProps,
		},
	})

	messageID := uuid.New().String()

	stream.Write(map[string]interface{}{
		"type":      "TEXT_MESSAGE_START",
		"messageId": messageID,
		"role":      "assistant",
	})

	collector := &messageCollector{stream: stream, messageID: messageID}
	if err := p.agent.Run(body, collector); err != nil {
		return
	}

	// Save assistant message
	p.storage.SaveMessage(body.ThreadID, body.RunID, messageID, "assistant", collector.content)

	stream.Write(map[string]interface{}{
		"type":      "TEXT_MESSAGE_END",
		"messageId": messageID,
	})

	stream.Write(map[string]interface{}{
		"type":     "RUN_FINISHED",
		"threadId": body.ThreadID,
		"runId":    body.RunID,
	})
}

type messageCollector struct {
	stream    SSEWriter
	messageID string
	content   string
}

func (m *messageCollector) Write(data interface{}) error {
	text := data.(string)
	m.content += text
	return m.stream.Write(map[string]interface{}{
		"type":      "TEXT_MESSAGE_CONTENT",
		"messageId": m.messageID,
		"delta":     text,
	})
}
