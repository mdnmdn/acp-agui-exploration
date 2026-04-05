package main

import (
	"context"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type GeminiAgent struct {
	llm       llms.Model
	verbose   bool
	streaming bool
}

func NewGeminiAgent(apiKey, model string, verbose, streaming bool) (*GeminiAgent, error) {
	opts := []googleai.Option{googleai.WithAPIKey(apiKey)}
	if model != "" {
		opts = append(opts, googleai.WithDefaultModel(model))
	}
	
	llm, err := googleai.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return &GeminiAgent{llm: llm, verbose: verbose, streaming: streaming}, nil
}

func (a *GeminiAgent) Info() map[string]interface{} {
	return map[string]interface{}{
		"version": "1.0.0",
		"agents": map[string]interface{}{
			"default": map[string]string{
				"name":        "default",
				"description": "Default AI agent",
				"className":   "GeminiAgent",
			},
		},
		"audioFileTranscriptionEnabled": false,
	}
}

func (a *GeminiAgent) Connect(body AgentBody) error {
	return nil
}

func (a *GeminiAgent) Run(body AgentBody, stream SSEWriter) error {
	for _, msg := range body.Messages {
		if a.verbose {
			log.Printf("[%s]: %s", msg.Role, msg.Content)
		}
	}

	ctx := context.Background()
	
	if a.verbose {
		log.Printf("Calling Gemini API (streaming=%v)", a.streaming)
	}

	if a.streaming {
		return a.runStreaming(ctx, body.Messages, stream)
	}
	return a.runNonStreaming(ctx, body.Messages, stream)
}

func (a *GeminiAgent) runStreaming(ctx context.Context, messages []Message, stream SSEWriter) error {
	response := ""
	_, err := a.llm.GenerateContent(ctx, a.buildMessages(messages), llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		text := string(chunk)
		response += text
		return stream.Write(text)
	}))

	if err != nil {
		log.Printf("Streaming error: %v", err)
		return err
	}

	if a.verbose {
		log.Printf("[assistant]: %s", response)
	}
	return nil
}

func (a *GeminiAgent) runNonStreaming(ctx context.Context, messages []Message, stream SSEWriter) error {
	result, err := a.llm.GenerateContent(ctx, a.buildMessages(messages))

	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}

	response := result.Choices[0].Content
	if a.verbose {
		log.Printf("[assistant]: %s", response)
	}

	for _, char := range response {
		if err := stream.Write(string(char)); err != nil {
			return err
		}
	}

	return nil
}

func (a *GeminiAgent) buildMessages(messages []Message) []llms.MessageContent {
	var result []llms.MessageContent
	for _, msg := range messages {
		role := llms.ChatMessageTypeHuman
		if msg.Role == "assistant" {
			role = llms.ChatMessageTypeAI
		}
		result = append(result, llms.MessageContent{
			Role: role,
			Parts: []llms.ContentPart{
				llms.TextPart(msg.Content),
			},
		})
	}
	return result
}
