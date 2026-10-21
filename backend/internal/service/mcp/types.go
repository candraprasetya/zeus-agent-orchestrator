package mcp

import "context"

// ServerConfig represents a discovered or configured MCP server.
type ServerConfig struct {
	Name       string            `json:"name"`
	Harness    string            `json:"harness"`
	Type       string            `json:"type"` // "local" | "remote"
	Command    string            `json:"command,omitempty"`
	Args       []string          `json:"args,omitempty"`
	URL        string            `json:"url,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	ConfigPath string            `json:"configPath,omitempty"`
	Enabled    bool              `json:"enabled"`
}

// AddServerRequest represents input to register a new MCP server.
type AddServerRequest struct {
	Name    string            `json:"name"`
	Harness string            `json:"harness"` // "opencode", "agy", "claude", etc.
	Type    string            `json:"type"`    // "local" or "remote"
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	URL     string            `json:"url,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Service manages MCP configurations across supported agent harnesses.
type Service interface {
	List(ctx context.Context, harness string) ([]ServerConfig, error)
	Add(ctx context.Context, req AddServerRequest) (ServerConfig, error)
	Delete(ctx context.Context, harness, name string) error
}
