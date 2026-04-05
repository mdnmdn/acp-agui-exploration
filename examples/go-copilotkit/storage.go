package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type StoredMessage struct {
	ID      string `json:"id"`
	RunID   string `json:"runId"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Thread struct {
	ID       string          `json:"id"`
	Messages []StoredMessage `json:"messages"`
}

type Storage interface {
	SaveMessage(threadID, runID, messageID, role, content string) error
	GetThread(threadID string) (*Thread, error)
}

type FileStorage struct {
	folder string
	mu     sync.RWMutex
}

func NewFileStorage(folder string) (*FileStorage, error) {
	if err := os.MkdirAll(folder, 0755); err != nil {
		return nil, err
	}
	return &FileStorage{folder: folder}, nil
}

func (s *FileStorage) threadPath(threadID string) string {
	return filepath.Join(s.folder, threadID+".json")
}

func (s *FileStorage) SaveMessage(threadID, runID, messageID, role, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	thread, _ := s.getThreadUnsafe(threadID)
	if thread == nil {
		thread = &Thread{ID: threadID, Messages: []StoredMessage{}}
	}

	thread.Messages = append(thread.Messages, StoredMessage{
		ID:      messageID,
		RunID:   runID,
		Role:    role,
		Content: content,
	})

	data, err := json.MarshalIndent(thread, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.threadPath(threadID), data, 0644)
}

func (s *FileStorage) GetThread(threadID string) (*Thread, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getThreadUnsafe(threadID)
}

func (s *FileStorage) getThreadUnsafe(threadID string) (*Thread, error) {
	data, err := os.ReadFile(s.threadPath(threadID))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var thread Thread
	if err := json.Unmarshal(data, &thread); err != nil {
		return nil, err
	}

	return &thread, nil
}
