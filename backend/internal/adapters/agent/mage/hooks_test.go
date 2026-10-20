package mage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestGetAgentHooksInstallsPluginAndSkill(t *testing.T) {
	workspace := t.TempDir()
	plugin := New()
	if err := plugin.GetAgentHooks(context.Background(), ports.WorkspaceHookConfig{WorkspacePath: workspace}); err != nil {
		t.Fatal(err)
	}

	pluginPath := magePluginPath(workspace)
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if !strings.Contains(source, magePluginSentinel) {
		t.Fatal("installed plugin is missing the AO ownership sentinel")
	}
	if !strings.Contains(source, mageHookCommandPrefix) {
		t.Fatalf("installed plugin never shells %q", mageHookCommandPrefix)
	}
	for _, event := range mageManagedEvents {
		if !strings.Contains(source, mageHookCommandPrefix+event) {
			t.Fatalf("installed plugin never reports %q", event)
		}
	}

	// Re-installing is idempotent and preserves AO ownership.
	if err := plugin.GetAgentHooks(context.Background(), ports.WorkspaceHookConfig{WorkspacePath: workspace}); err != nil {
		t.Fatal(err)
	}
	installed, err := plugin.AreHooksInstalled(context.Background(), workspace)
	if err != nil {
		t.Fatal(err)
	}
	if !installed {
		t.Fatal("AreHooksInstalled = false after install")
	}

	if err := plugin.UninstallHooks(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
		t.Fatalf("plugin still present after uninstall: %v", err)
	}
	installed, err = plugin.AreHooksInstalled(context.Background(), workspace)
	if err != nil {
		t.Fatal(err)
	}
	if installed {
		t.Fatal("AreHooksInstalled = true after uninstall")
	}
}

func TestGetAgentHooksRefusesForeignPluginFile(t *testing.T) {
	workspace := t.TempDir()
	pluginPath := magePluginPath(workspace)
	if err := os.MkdirAll(filepath.Dir(pluginPath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pluginPath, []byte("export default async function () {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := New().GetAgentHooks(context.Background(), ports.WorkspaceHookConfig{WorkspacePath: workspace})
	if err == nil {
		t.Fatal("GetAgentHooks succeeded over a user-authored file at the AO plugin path")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("error = %v, want a refusing-to-overwrite diagnostic", err)
	}
}
