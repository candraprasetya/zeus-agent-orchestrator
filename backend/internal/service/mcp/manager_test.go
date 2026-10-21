package mcp

import (
	"context"
	"testing"
)

func TestMCPManagerOpenCode(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := New(tmpDir)
	ctx := context.Background()

	// Initially empty
	list, err := mgr.List(ctx, "opencode")
	if err != nil {
		t.Fatalf("unexpected error on empty list: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 servers, got %d", len(list))
	}

	// Add local server
	addReq := AddServerRequest{
		Name:    "git-mcp",
		Harness: "opencode",
		Type:    "local",
		Command: "npx",
		Args:    []string{"-y", "mcp-server-git"},
		Env:     map[string]string{"GIT_USER": "test"},
	}
	added, err := mgr.Add(ctx, addReq)
	if err != nil {
		t.Fatalf("failed to add server: %v", err)
	}
	if added.Name != "git-mcp" || added.Command != "npx" || len(added.Args) != 2 {
		t.Fatalf("unexpected added server: %+v", added)
	}

	// Add remote server
	remoteReq := AddServerRequest{
		Name:    "figma-remote",
		Harness: "opencode",
		Type:    "remote",
		URL:     "https://mcp.figma.com/mcp",
	}
	_, err = mgr.Add(ctx, remoteReq)
	if err != nil {
		t.Fatalf("failed to add remote server: %v", err)
	}

	// Verify list has 2 items
	list, err = mgr.List(ctx, "opencode")
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(list))
	}

	// Delete git-mcp
	if err := mgr.Delete(ctx, "opencode", "git-mcp"); err != nil {
		t.Fatalf("failed to delete server: %v", err)
	}

	list, err = mgr.List(ctx, "opencode")
	if err != nil {
		t.Fatalf("failed to list after delete: %v", err)
	}
	if len(list) != 1 || list[0].Name != "figma-remote" {
		t.Fatalf("expected 1 remaining server (figma-remote), got %+v", list)
	}
}

func TestMCPManagerAgyAndClaude(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := New(tmpDir)
	ctx := context.Background()

	// Add AGY server
	_, err := mgr.Add(ctx, AddServerRequest{
		Name:    "agy-test",
		Harness: "agy",
		Type:    "local",
		Command: "node",
		Args:    []string{"server.js"},
	})
	if err != nil {
		t.Fatalf("failed to add agy server: %v", err)
	}

	agyList, err := mgr.List(ctx, "agy")
	if err != nil || len(agyList) != 1 {
		t.Fatalf("expected 1 agy server, got %v (%d)", err, len(agyList))
	}

	// Add Claude server
	_, err = mgr.Add(ctx, AddServerRequest{
		Name:    "claude-test",
		Harness: "claude",
		Type:    "local",
		Command: "uvx",
		Args:    []string{"mcp-server"},
	})
	if err != nil {
		t.Fatalf("failed to add claude server: %v", err)
	}

	claudeList, err := mgr.List(ctx, "claude")
	if err != nil || len(claudeList) != 1 {
		t.Fatalf("expected 1 claude server, got %v (%d)", err, len(claudeList))
	}

	// Add Mage server
	_, err = mgr.Add(ctx, AddServerRequest{
		Name:    "mage-test",
		Harness: "mage",
		Type:    "local",
		Command: "node",
		Args:    []string{"mage-mcp.js"},
	})
	if err != nil {
		t.Fatalf("failed to add mage server: %v", err)
	}

	mageList, err := mgr.List(ctx, "mage")
	if err != nil || len(mageList) != 1 {
		t.Fatalf("expected 1 mage server, got %v (%d)", err, len(mageList))
	}

	// List all
	allList, err := mgr.List(ctx, "all")
	if err != nil || len(allList) != 3 {
		t.Fatalf("expected 3 total servers, got %v (%d)", err, len(allList))
	}
}

func TestMCPManagerOpenCodeNestedServers(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := New(tmpDir)
	ctx := context.Background()

	// Add a server under opencode-v2
	_, err := mgr.Add(ctx, AddServerRequest{
		Name:    "figcraft",
		Harness: "opencode-v2",
		Type:    "local",
		Command: "figcraft-mcp",
	})
	if err != nil {
		t.Fatalf("failed to add figcraft: %v", err)
	}

	list, err := mgr.List(ctx, "opencode-v2")
	if err != nil {
		t.Fatalf("failed to list opencode-v2: %v", err)
	}
	if len(list) != 1 || list[0].Name != "figcraft" || list[0].Command != "figcraft-mcp" {
		t.Fatalf("unexpected list: %+v", list)
	}

	// Delete
	if err := mgr.Delete(ctx, "opencode-v2", "figcraft"); err != nil {
		t.Fatalf("failed to delete figcraft: %v", err)
	}

	list, err = mgr.List(ctx, "opencode-v2")
	if err != nil {
		t.Fatalf("failed to list after delete: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 servers after delete, got %d", len(list))
	}
}
