package mage

import "github.com/aoagents/agent-orchestrator/backend/internal/domain"

// DeriveActivityState maps a mage plugin hook event onto an AO activity state.
// The mage plugin (assets/ao-activity.ts) normalizes mage's native events to
// "session-start" / "user-prompt-submit" / "stop" before invoking
// `ao hooks mage <event>`. The bool is false when the event carries no
// activity signal.
func DeriveActivityState(event string, _ []byte) (domain.ActivityState, bool) {
	switch event {
	case "session-start":
		return domain.ActivityActive, true
	case "user-prompt-submit":
		return domain.ActivityActive, true
	case "active":
		return domain.ActivityActive, true
	case "stop":
		return domain.ActivityIdle, true
	case "permission-blocked":
		return domain.ActivityBlocked, true
	default:
		return "", false
	}
}
