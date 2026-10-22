// Package mageacp binds the user's own mage installation to AO's reusable ACP
// Chat transport.
package mageacp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	acpsdk "github.com/coder/acp-go-sdk"

	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/mage"
	acpdriver "github.com/aoagents/agent-orchestrator/backend/internal/adapters/chatdriver/acp"
	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/chatdriver/nativeacp"
	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/chatdriver/opencodeidentity"
	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

// New launches `mage acp` from the exact binary resolved by the existing mage
// agent plugin. AO adds only a per-session inline overlay for its standing
// instructions and an explicit bypass-permissions choice, keyed on mage's
// MAGE_CONFIG_CONTENT overlay (its analogue of OPENCODE_CONFIG_CONTENT).
func New(plugin nativeacp.Plugin, log *slog.Logger) ports.ChatDriver {
	return nativeacp.New(plugin, nativeacp.Config{
		Harness:              domain.HarnessMage,
		Configure:            configure,
		SessionOptions:       sessionOptions,
		PermissionPolicy:     permissionPolicy,
		ValidateTurnSettings: validateTurnSettings,
		DecodeProviderConversationID: func(providerID string) (string, error) {
			return opencodeidentity.Decode(domain.HarnessMage, providerID)
		},
	}, log)
}

func configure(_ context.Context, cfg acpdriver.LaunchConfig) ([]string, map[string]string, error) {
	content, err := mage.PrepareACPConfigContent(
		cfg.Env["MAGE_CONFIG_CONTENT"], cfg.SystemPrompt, cfg.Permissions)
	if err != nil {
		return nil, nil, err
	}
	return []string{"acp", "--pure"}, map[string]string{"MAGE_CONFIG_CONTENT": content}, nil
}

func sessionOptions(settings ports.ChatTurnSettings) []acpdriver.SessionOption {
	options := make([]acpdriver.SessionOption, 0, 2)
	if settings.Model != "" {
		options = append(options, acpdriver.SessionOption{ID: "model", Value: settings.Model})
	}
	// mage advertises effort as id "effort" (category "thought_level") with the
	// model's variant names as values, inherited from opencode. The model setter
	// resets the variant to "default" else the first variant when no explicit
	// variant is given, so the model must be applied first and the effort second.
	if settings.Effort != "" {
		options = append(options, acpdriver.SessionOption{ID: "effort", Value: settings.Effort})
	}
	if len(options) == 0 {
		return nil
	}
	return options
}

// mage parses model overrides as provider/model. A provider display name is
// not an alias for its configured default model; keep that default only when
// the override is empty. Leave model availability to the user's mage.
func validateTurnSettings(_ ports.PermissionMode, settings ports.ChatTurnSettings) error {
	if settings.Model == "" {
		return nil
	}
	provider, model, found := strings.Cut(settings.Model, "/")
	if !found || strings.TrimSpace(provider) == "" || strings.TrimSpace(model) == "" {
		return fmt.Errorf("%w: mage model %q must use provider/model format (for example, merlin/default); select a full model ID from `mage models`, or clear the model override to use agent settings", ports.ErrChatConfigOptionInvalid, settings.Model)
	}
	return nil
}

// permissionPolicy is AO's side of accept-edits and auto, and mirrors how mage
// implements --auto: reply to each permission request rather than granting
// anything up front. A tool the user or repository denied never raises a
// request, so provider policy stays authoritative and AO never has to
// reconstruct it. Bypass needs no entry — its agent allows everything, so
// nothing is asked.
func permissionPolicy(
	mode ports.PermissionMode,
	params acpsdk.RequestPermissionRequest,
) (acpsdk.PermissionOptionId, bool) {
	mode = ports.NormalizePermissionMode(mode)
	kind := params.ToolCall.Kind
	if mode != ports.PermissionModeAuto && (mode != ports.PermissionModeAcceptEdits || kind == nil ||
		(*kind != acpsdk.ToolKindEdit && *kind != acpsdk.ToolKindDelete && *kind != acpsdk.ToolKindMove)) {
		return "", false
	}
	// Once, not always: --auto answers each request and persists nothing.
	for _, option := range params.Options {
		if option.Kind == acpsdk.PermissionOptionKindAllowOnce {
			return option.OptionId, true
		}
	}
	return "", false
}
