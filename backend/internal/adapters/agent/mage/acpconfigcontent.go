package mage

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

// prepareACPConfigContent merges AO's standing instructions and its approval
// tiers into mage's inline runtime overlay (MAGE_CONFIG_CONTENT). The user's
// MAGE_CONFIG path remains untouched, preserving its normal global, custom,
// project, provider, and credential configuration. mage accepts the
// https://opencode.ai/config.json schema, so the overlay shape matches the
// opencode adapter's.
func prepareACPConfigContent(
	existing, systemPrompt string,
	permissions ports.PermissionMode,
) (string, error) {
	config := map[string]any{}
	if strings.TrimSpace(existing) != "" {
		if err := json.Unmarshal([]byte(existing), &config); err != nil {
			return "", fmt.Errorf("decode MAGE_CONFIG_CONTENT: %w", err)
		}
	}
	if _, ok := config["$schema"]; !ok {
		config["$schema"] = "https://opencode.ai/config.json"
	}
	agents, ok := config["agent"].(map[string]any)
	if config["agent"] != nil && !ok {
		return "", fmt.Errorf("MAGE_CONFIG_CONTENT agent must be an object")
	}
	if agents == nil {
		agents = map[string]any{}
	}
	for _, mode := range []ports.PermissionMode{
		ports.PermissionModeDefault, ports.PermissionModeAcceptEdits,
		ports.PermissionModeAuto, ports.PermissionModeBypassPermissions,
	} {
		agent := mageAgentSettings{Mode: "primary", Prompt: systemPrompt}
		// Only bypass writes a rule. Accept edits and auto answer requests,
		// like mage's --auto, so explicit provider denies stay denied.
		if mode == ports.PermissionModeBypassPermissions {
			agent.Permission = "allow"
		}
		agents[acpAgentForPermissions(mode)] = agent
	}
	config["agent"] = agents
	config["default_agent"] = acpAgentForPermissions(permissions)
	data, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("encode ACP agent config: %w", err)
	}
	return string(data), nil
}

// acpAgentForPermissions names the per-approval-tier agent the ACP config
// overlay defines, so every tier stays selectable inside one configured agent.
func acpAgentForPermissions(permissions ports.PermissionMode) string {
	switch ports.NormalizePermissionMode(permissions) {
	case ports.PermissionModeAcceptEdits:
		return "ao-accept-edits"
	case ports.PermissionModeAuto:
		return "ao-auto"
	case ports.PermissionModeBypassPermissions:
		return "ao-bypass"
	default:
		return "ao-default"
	}
}
