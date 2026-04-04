package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

const (
	registryURL = "https://cdn.agentclientprotocol.com/registry/v1/latest/registry.json"
	cacheTTL     = 1 * time.Hour
)

type Registry struct {
	Version   string  `json:"version"`
	Agents    []Agent `json:"agents"`
	Extensions []any   `json:"extensions"`
}

type Agent struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Version      string     `json:"version"`
	Description  string     `json:"description"`
	Repository   string     `json:"repository,omitempty"`
	Website      string     `json:"website,omitempty"`
	Authors      []string   `json:"authors"`
	License      string     `json:"license"`
	Icon         string     `json:"icon,omitempty"`
	Distribution struct {
		Binary map[string]BinaryTarget `json:"binary,omitempty"`
		Npx    *NpxTarget              `json:"npx,omitempty"`
		Uvx    *UvxTarget              `json:"uvx,omitempty"`
	} `json:"distribution"`
}

type BinaryTarget struct {
	Archive string            `json:"archive"`
	Cmd     string            `json:"cmd"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

type NpxTarget struct {
	Package string   `json:"package"`
	Args    []string `json:"args,omitempty"`
}

type UvxTarget struct {
	Package string   `json:"package"`
	Args    []string `json:"args,omitempty"`
}

type RegistryCache struct {
	cache      *Registry
	lastFetch  time.Time
	cacheFile  string
	mu         sync.Mutex
}

func NewRegistryCache() *RegistryCache {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	cacheFile := cacheDir + "/acp-registry-cache.json"
	return &RegistryCache{
		cacheFile: cacheFile,
	}
}

func (rc *RegistryCache) Get() (*Registry, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Try to load from cache first
	if rc.cache != nil && time.Since(rc.lastFetch) < cacheTTL {
		return rc.cache, nil
	}

	// Try to load from file cache
	if fileCache, err := rc.loadFromFile(); err == nil {
		rc.cache = fileCache
		rc.lastFetch = time.Now()
		return rc.cache, nil
	}

	// Fetch fresh from network
	registry, err := fetchRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry: %w", err)
	}

	// Save to file cache
	if err := rc.saveToFile(registry); err != nil {
		fmt.Printf("Warning: failed to save registry cache: %v\n", err)
	}

	rc.cache = registry
	rc.lastFetch = time.Now()
	return registry, nil
}

func (rc *RegistryCache) loadFromFile() (*Registry, error) {
	data, err := os.ReadFile(rc.cacheFile)
	if err != nil {
		return nil, err
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	return &registry, nil
}

func (rc *RegistryCache) saveToFile(registry *Registry) error {
	data, err := json.Marshal(registry)
	if err != nil {
		return err
	}

	return os.WriteFile(rc.cacheFile, data, 0644)
}

func fetchRegistry() (*Registry, error) {
	resp, err := http.Get(registryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry response: %w", err)
	}

	var registry Registry
	if err := json.Unmarshal(body, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	return &registry, nil
}

func getCurrentPlatform() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "darwin":
		if arch == "arm64" {
			return "darwin-aarch64"
		}
		return "darwin-x86_64"
	case "linux":
		if arch == "arm64" {
			return "linux-aarch64"
		}
		return "linux-x86_64"
	case "windows":
		if arch == "arm64" {
			return "windows-aarch64"
		}
		return "windows-x86_64"
	default:
		return "linux-x86_64"
	}
}

func getCompatibleAgents(registry *Registry) []Agent {
	platform := getCurrentPlatform()
	var compatible []Agent

	for _, agent := range registry.Agents {
		if agent.Distribution.Binary != nil {
			if _, exists := agent.Distribution.Binary[platform]; exists {
				compatible = append(compatible, agent)
				continue
			}
		}
		if agent.Distribution.Npx != nil {
			compatible = append(compatible, agent)
			continue
		}
		if agent.Distribution.Uvx != nil {
			compatible = append(compatible, agent)
		}
	}

	return compatible
}

func agentToAgentInfo(agent Agent) agentInfo {
	var command string
	var args []string

	platform := getCurrentPlatform()

	// Determine command based on distribution type
	if agent.Distribution.Binary != nil {
		if target, exists := agent.Distribution.Binary[platform]; exists {
			command = target.Cmd
			args = target.Args
			if len(args) == 0 {
				args = []string{"--acp"}
			}
		}
	} else if agent.Distribution.Npx != nil {
		command = "npx"
		args = append([]string{agent.Distribution.Npx.Package}, agent.Distribution.Npx.Args...)
	} else if agent.Distribution.Uvx != nil {
		command = "uvx"
		args = append([]string{agent.Distribution.Uvx.Package}, agent.Distribution.Uvx.Args...)
	}

	return agentInfo{
		name:    strings.ToLower(agent.ID),
		command: command,
		args:    args,
		desc:    fmt.Sprintf("%s - %s", agent.Name, agent.Description),
	}
}

func LoadAgentsFromRegistry() ([]list.Item, error) {
	cache := NewRegistryCache()
	registry, err := cache.Get()
	if err != nil {
		fmt.Printf("Warning: Failed to load registry, using built-in agents: %v\n", err)
		return agents, nil
	}

	compatible := getCompatibleAgents(registry)
	if len(compatible) == 0 {
		fmt.Println("Warning: No compatible agents found in registry, using built-in agents")
		return agents, nil
	}

	var registryAgents []list.Item
	for _, agent := range compatible {
		registryAgents = append(registryAgents, agentToAgentInfo(agent))
	}

	return registryAgents, nil
}
