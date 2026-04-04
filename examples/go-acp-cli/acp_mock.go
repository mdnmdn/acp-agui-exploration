package main

// Mock ACP types and functions to allow the code to compile
// This is a minimal implementation that provides just what's needed

import (
	"context"
)

// Mock types
type ClientSideConnection struct {
	done chan struct{}
}

type AgentSideConnection struct{}

type InitializeRequest struct {
	ClientCapabilities ClientCapabilities
	ClientInfo         *Implementation
	ProtocolVersion    string
}

type ClientCapabilities struct{}

type Implementation struct {
	Name    string
	Version string
}

type NewSessionRequest struct {
	Cwd        string
	McpServers []McpServer
}

type NewSessionResponse struct {
	SessionId string
}

type McpServer struct{}

type PromptRequest struct {
	SessionId string
	Prompt    []ContentBlock
}

type PromptResponse struct{}

type ContentBlock struct {
	Text *TextBlock
}

type TextBlock struct {
	Text string
}

type ReadTextFileRequest struct{}
type ReadTextFileResponse struct{}

type WriteTextFileRequest struct{}
type WriteTextFileResponse struct{}

type RequestPermissionRequest struct{}
type RequestPermissionResponse struct{}

type SessionNotification struct {
	Update SessionUpdate
}

type SessionUpdate struct {
	AgentMessageChunk *SessionUpdateAgentMessageChunk
}

type SessionUpdateAgentMessageChunk struct {
	Content ContentBlock
}

type CreateTerminalRequest struct{}
type CreateTerminalResponse struct{}

type KillTerminalCommandRequest struct{}
type KillTerminalCommandResponse struct{}

type TerminalOutputRequest struct{}
type TerminalOutputResponse struct{}

type ReleaseTerminalRequest struct{}
type ReleaseTerminalResponse struct{}

type WaitForTerminalExitRequest struct{}
type WaitForTerminalExitResponse struct{}

type AuthenticateRequest struct{}
type AuthenticateResponse struct{}

type CancelNotification struct{}

type SetSessionModeRequest struct{}
type SetSessionModeResponse struct{}

type SetSessionModelRequest struct {
	SessionId string
	ModelId   ModelId
}

type SetSessionModelResponse struct{}

type ModelId string
type SessionId string

// Mock functions
func NewClientSideConnection(client interface{}, stdin, stdout interface{}) *ClientSideConnection {
	return &ClientSideConnection{done: make(chan struct{})}
}

func NewAgentSideConnection(agent interface{}, stdout, stderr interface{}) *AgentSideConnection {
	return &AgentSideConnection{}
}

func (c *ClientSideConnection) Initialize(ctx context.Context, req InitializeRequest) (InitializeResponse, error) {
	return InitializeResponse{}, nil
}

func (c *ClientSideConnection) NewSession(ctx context.Context, req NewSessionRequest) (NewSessionResponse, error) {
	return NewSessionResponse{SessionId: "mock-session"}, nil
}

func (c *ClientSideConnection) Prompt(ctx context.Context, req PromptRequest) (PromptResponse, error) {
	return PromptResponse{}, nil
}

func (c *ClientSideConnection) SetSessionModel(ctx context.Context, req SetSessionModelRequest) (SetSessionModelResponse, error) {
	return SetSessionModelResponse{}, nil
}

func (c *ClientSideConnection) ReadTextFile(ctx context.Context, params ReadTextFileRequest) (ReadTextFileResponse, error) {
	return ReadTextFileResponse{}, nil
}

func (c *ClientSideConnection) WriteTextFile(ctx context.Context, params WriteTextFileRequest) (WriteTextFileResponse, error) {
	return WriteTextFileResponse{}, nil
}

func (c *ClientSideConnection) RequestPermission(ctx context.Context, params RequestPermissionRequest) (RequestPermissionResponse, error) {
	return RequestPermissionResponse{}, nil
}

func (c *ClientSideConnection) SessionUpdate(ctx context.Context, params SessionNotification) error {
	return nil
}

func (c *ClientSideConnection) CreateTerminal(ctx context.Context, params CreateTerminalRequest) (CreateTerminalResponse, error) {
	return CreateTerminalResponse{}, nil
}

func (c *ClientSideConnection) KillTerminalCommand(ctx context.Context, params KillTerminalCommandRequest) (KillTerminalCommandResponse, error) {
	return KillTerminalCommandResponse{}, nil
}

func (c *ClientSideConnection) TerminalOutput(ctx context.Context, params TerminalOutputRequest) (TerminalOutputResponse, error) {
	return TerminalOutputResponse{}, nil
}

func (c *ClientSideConnection) ReleaseTerminal(ctx context.Context, params ReleaseTerminalRequest) (ReleaseTerminalResponse, error) {
	return ReleaseTerminalResponse{}, nil
}

func (c *ClientSideConnection) WaitForTerminalExit(ctx context.Context, params WaitForTerminalExitRequest) (WaitForTerminalExitResponse, error) {
	return WaitForTerminalExitResponse{}, nil
}

func (c *ClientSideConnection) Done() <-chan struct{} {
	return nil
}

func (a *AgentSideConnection) Authenticate(ctx context.Context, params AuthenticateRequest) (AuthenticateResponse, error) {
	return AuthenticateResponse{}, nil
}

func (a *AgentSideConnection) Initialize(ctx context.Context, params InitializeRequest) (InitializeResponse, error) {
	return InitializeResponse{}, nil
}

func (a *AgentSideConnection) Cancel(ctx context.Context, params CancelNotification) error {
	return nil
}

func (a *AgentSideConnection) NewSession(ctx context.Context, params NewSessionRequest) (NewSessionResponse, error) {
	return NewSessionResponse{SessionId: "mock-session"}, nil
}

func (a *AgentSideConnection) Prompt(ctx context.Context, params PromptRequest) (PromptResponse, error) {
	return PromptResponse{}, nil
}

func (a *AgentSideConnection) SetSessionMode(ctx context.Context, params SetSessionModeRequest) (SetSessionModeResponse, error) {
	return SetSessionModeResponse{}, nil
}

func (a *AgentSideConnection) SetSessionModel(ctx context.Context, params SetSessionModelRequest) (SetSessionModelResponse, error) {
	return SetSessionModelResponse{}, nil
}

func (conn *AgentSideConnection) SessionUpdate(ctx context.Context, params SessionNotification) error {
	return nil
}

type InitializeResponse struct{}

// Mock ACP package variables - using string directly
const ProtocolVersionNumber = "0.1.0"
