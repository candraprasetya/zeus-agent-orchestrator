package mage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

// fakeMageBinary installs a stub mage executable on PATH so launch/restore
// argv construction is deterministic without probing the real install.
func fakeMageBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	binary := filepath.Join(dir, "mage")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\necho mage 1.0.0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	return binary
}

func TestGetLaunchCommandBuildsArgv(t *testing.T) {
	binary := fakeMageBinary(t)
	plugin := New()
	promptFile := filepath.Join(t.TempDir(), "system.md")

	cmd, err := plugin.GetLaunchCommand(context.Background(), ports.LaunchConfig{
		Permissions:      ports.PermissionModeBypassPermissions,
		Prompt:           "-fix this",
		SessionID:        "sess/1",
		SystemPromptFile: promptFile,
		SystemPrompt:     "follow AO rules",
	})
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(filepath.Dir(promptFile), "mage.json")
	// Bypass carries no launch flag: mage has no
	// --dangerously-skip-permissions alias, so the full-access rule rides on
	// the AO agent instead.
	want := []string{
		"env", "MAGE_CONFIG=" + configPath,
		binary,
		"--agent", "ao-sess-1",
		"--prompt", "-fix this",
	}
	if !reflect.DeepEqual(cmd, want) {
		t.Fatalf("unexpected command\nwant: %#v\n got: %#v", want, cmd)
	}
	var config mageInlineConfig
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	agent := config.Agent["ao-sess-1"]
	if agent.Mode != "primary" || agent.Prompt != "follow AO rules" {
		t.Fatalf("agent config = %#v, want primary inline prompt", agent)
	}
	if agent.Permission != "allow" {
		t.Fatalf("agent permission = %#v, want allow (bypass outranks every config layer)", agent.Permission)
	}
	if config.Permission != nil {
		t.Fatalf("top-level permission = %#v, want nil when bypass rides on the agent", config.Permission)
	}
}

func TestGetLaunchCommandSystemPromptFileConfig(t *testing.T) {
	binary := fakeMageBinary(t)
	plugin := New()
	promptFile := filepath.Join(t.TempDir(), "system.md")

	cmd, err := plugin.GetLaunchCommand(context.Background(), ports.LaunchConfig{
		SessionID:        "sess-2",
		SystemPromptFile: promptFile,
	})
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(filepath.Dir(promptFile), "mage.json")
	want := []string{"env", "MAGE_CONFIG=" + configPath, binary, "--agent", "ao-sess-2"}
	if !reflect.DeepEqual(cmd, want) {
		t.Fatalf("unexpected command\nwant: %#v\n got: %#v", want, cmd)
	}
	var config mageInlineConfig
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	if got := config.Agent["ao-sess-2"].Prompt; got != "{file:./system.md}" {
		t.Fatalf("agent prompt = %q, want file reference", got)
	}
}

func TestGetLaunchCommandMapsPermissionModes(t *testing.T) {
	fakeMageBinary(t)
	tests := []struct {
		name       string
		permission ports.PermissionMode
		want       map[string]string
		wantFlag   string
	}{
		{name: "default", permission: ports.PermissionModeDefault},
		{name: "accept-edits", permission: ports.PermissionModeAcceptEdits, want: map[string]string{"edit": "allow"}},
		{name: "auto", permission: ports.PermissionModeAuto, wantFlag: "--auto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptFile := filepath.Join(t.TempDir(), "system.md")
			cmd, err := New().GetLaunchCommand(context.Background(), ports.LaunchConfig{
				Permissions:      tt.permission,
				SessionID:        "s",
				SystemPromptFile: promptFile,
			})
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(filepath.Dir(promptFile), "mage.json"))
			if err != nil {
				t.Fatal(err)
			}
			var config mageInlineConfig
			if err := json.Unmarshal(data, &config); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(config.Permission, tt.want) {
				t.Fatalf("permission = %#v, want %#v", config.Permission, tt.want)
			}
			flag := ""
			for _, arg := range cmd {
				if arg == "--auto" {
					flag = arg
				}
			}
			if flag != tt.wantFlag {
				t.Fatalf("flag = %q, want %q", flag, tt.wantFlag)
			}
		})
	}
}

func TestGetLaunchCommandWithoutSystemPromptHasNoOverlay(t *testing.T) {
	binary := fakeMageBinary(t)
	cmd, err := New().GetLaunchCommand(context.Background(), ports.LaunchConfig{
		Permissions: ports.PermissionModeAuto,
		Prompt:      "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{binary, "--auto", "--prompt", "hello"}
	if !reflect.DeepEqual(cmd, want) {
		t.Fatalf("unexpected command\nwant: %#v\n got: %#v", want, cmd)
	}
}

func TestGetRestoreCommandUsesNativeSessionID(t *testing.T) {
	binary := fakeMageBinary(t)
	promptFile := filepath.Join(t.TempDir(), "system.md")
	cmd, ok, err := New().GetRestoreCommand(context.Background(), ports.RestoreConfig{
		Session: ports.SessionRef{
			ID:       "sess-3",
			Metadata: map[string]string{"agentSessionId": "native-123"},
		},
		SystemPromptFile: promptFile,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ok = false, want restore with native session id")
	}
	configPath := filepath.Join(filepath.Dir(promptFile), "mage.json")
	want := []string{
		"env", "MAGE_CONFIG=" + configPath,
		binary,
		"--agent", "ao-sess-3",
		"--session", "native-123",
	}
	if !reflect.DeepEqual(cmd, want) {
		t.Fatalf("unexpected command\nwant: %#v\n got: %#v", want, cmd)
	}
}

func TestGetRestoreCommandFallsBackWithoutNativeSessionID(t *testing.T) {
	fakeMageBinary(t)
	_, ok, err := New().GetRestoreCommand(context.Background(), ports.RestoreConfig{
		Session: ports.SessionRef{ID: "sess-4"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("ok = true, want false without a native session id")
	}
}

func TestPrepareACPConfigContentDefinesEveryPermissionAgent(t *testing.T) {
	content, err := PrepareACPConfigContent("", "AO standing instructions", ports.PermissionModeBypassPermissions)
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Schema       string         `json:"$schema"`
		Agents       map[string]any `json:"agent"`
		DefaultAgent string         `json:"default_agent"`
	}
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ao-default", "ao-accept-edits", "ao-auto", "ao-bypass"} {
		if _, ok := config.Agents[name]; !ok {
			t.Fatalf("agent %q missing from ACP overlay", name)
		}
	}
	if config.DefaultAgent != "ao-bypass" {
		t.Fatalf("default_agent = %q, want ao-bypass", config.DefaultAgent)
	}
	if config.Schema != "https://opencode.ai/config.json" {
		t.Fatalf("schema = %q, want the opencode-compatible schema mage reads", config.Schema)
	}
}

func TestAuthStatusUnknownWhenNoCredentials(t *testing.T) {
	// mage's own store is empty on a fresh machine; the outcome must stay
	// unknown so free models remain launchable.
	t.Setenv("MAGE_DATA_DIR", t.TempDir())
	for _, key := range mageAPIKeyEnvVars {
		t.Setenv(key, "")
	}
	status, err := New().AuthStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status != ports.AgentAuthStatusUnknown {
		t.Fatalf("status = %q, want unknown", status)
	}
}

func TestAuthStatusAuthorizedWithStoredCredentialFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MAGE_DATA_DIR", dir)
	for _, key := range mageAPIKeyEnvVars {
		t.Setenv(key, "")
	}
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{"merlin":{"type":"api","key":"secret"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := New().AuthStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status != ports.AgentAuthStatusAuthorized {
		t.Fatalf("status = %q, want authorized", status)
	}
}
