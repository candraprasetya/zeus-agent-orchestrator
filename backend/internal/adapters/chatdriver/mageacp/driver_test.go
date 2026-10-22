package mageacp

import (
	"context"
	"errors"
	"strings"
	"testing"

	acpsdk "github.com/coder/acp-go-sdk"

	acpdriver "github.com/aoagents/agent-orchestrator/backend/internal/adapters/chatdriver/acp"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestConfigureUsesAcpPureArgs(t *testing.T) {
	args, env, err := configure(context.Background(), acpdriver.LaunchConfig{
		SessionID: "worker-1", SystemPrompt: "Test prompt", Permissions: ports.PermissionModeDefault,
	})
	if err != nil {
		t.Fatalf("configure: %v", err)
	}
	if len(args) != 2 || args[0] != "acp" || args[1] != "--pure" {
		t.Fatalf("args = %#v, want [\"acp\", \"--pure\"]", args)
	}
	if env["MAGE_CONFIG_CONTENT"] == "" {
		t.Fatalf("MAGE_CONFIG_CONTENT is empty in env")
	}
}

func TestSessionOptionsUseProviderAdvertisedModelOption(t *testing.T) {
	if got := sessionOptions(ports.ChatTurnSettings{}); got != nil {
		t.Fatalf("empty settings = %#v", got)
	}
	got := sessionOptions(ports.ChatTurnSettings{Model: "merlin/default"})
	if len(got) != 1 || got[0].ID != "model" || got[0].Value != "merlin/default" {
		t.Fatalf("model settings = %#v", got)
	}
}

func TestRejectsInvalidModelFormatBeforeLaunchingMage(t *testing.T) {
	plugin := &unresolvedPlugin{}
	driver := New(plugin, nil)
	cfg := ports.ChatStartConfig{WorkspacePath: t.TempDir(), Model: "invalid-bare-model"}
	_, err := driver.Start(context.Background(), cfg)
	if !errors.Is(err, ports.ErrChatConfigOptionInvalid) || !strings.Contains(err.Error(), "provider/model") {
		t.Fatalf("error = %v, want model format validation with recovery guidance", err)
	}
	if plugin.resolved {
		t.Fatal("invalid model reached Mage binary resolution")
	}
}

type unresolvedPlugin struct{ resolved bool }

func (p *unresolvedPlugin) ResolveBinary(context.Context) (string, error) {
	p.resolved = true
	return "", errors.New("unexpected binary resolution")
}
func (*unresolvedPlugin) AuthStatus(context.Context) (ports.AgentAuthStatus, error) {
	return ports.AgentAuthStatusUnknown, nil
}

func TestValidateTurnSettingsModelFormat(t *testing.T) {
	for _, model := range []string{"", "merlin/default", "custom/model"} {
		t.Run("valid/"+model, func(t *testing.T) {
			settings := ports.ChatTurnSettings{Model: model}
			if err := validateTurnSettings(ports.PermissionModeDefault, settings); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, model := range []string{"merlin", "default", " ", "/model", "provider/", " /model", "provider/ "} {
		t.Run("invalid/"+model, func(t *testing.T) {
			err := validateTurnSettings(ports.PermissionModeDefault, ports.ChatTurnSettings{Model: model})
			if !errors.Is(err, ports.ErrChatConfigOptionInvalid) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestPermissionPolicyAnswersRequests(t *testing.T) {
	edit, execute := acpsdk.ToolKindEdit, acpsdk.ToolKindExecute
	options := []acpsdk.PermissionOption{
		{OptionId: "reject", Kind: acpsdk.PermissionOptionKindRejectOnce},
		{OptionId: "once", Kind: acpsdk.PermissionOptionKindAllowOnce},
	}
	id, handled := permissionPolicy(ports.PermissionModeAcceptEdits, acpsdk.RequestPermissionRequest{
		ToolCall: acpsdk.ToolCallUpdate{Kind: &edit}, Options: options,
	})
	if !handled || id != "once" {
		t.Fatalf("expected handled once for edit, got id=%s, handled=%v", id, handled)
	}

	_, handled = permissionPolicy(ports.PermissionModeAcceptEdits, acpsdk.RequestPermissionRequest{
		ToolCall: acpsdk.ToolCallUpdate{Kind: &execute}, Options: options,
	})
	if handled {
		t.Fatal("expected execute to be unhandled in accept-edits mode")
	}
}
