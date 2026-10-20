// Package mage implements the mage agent adapter: launching new TUI sessions,
// resuming sessions by native id, installing a workspace-local activity plugin
// plus the using-ao skill, and reading plugin-derived session info.
//
// mage is an opencode-compatible fork (same https://opencode.ai/config.json
// schema, same `{plugin,plugins}/*.{ts,js}` plugin glob, same native flags
// --model/--agent/--prompt/--session), so AO bridges it the same way as
// opencode:
//   - It has no native command-hook config. Its only lifecycle-extensibility
//     surface is a JS/TS plugin loaded from .mage/plugins/, so GetAgentHooks
//     installs an AO-owned plugin file (see hooks.go) instead of merging JSON.
//     The same install also materializes using-ao under .mage/skills/ so
//     mage's skill discovery can find it.
//   - Its CLI exposes no system-prompt flag, so AO injects standing
//     instructions by writing an AO-owned per-session config (MAGE_CONFIG)
//     and selecting the generated agent.
//
// AO-managed sessions derive native session identity and display metadata
// from the mage plugin's reported events, mirroring the opencode adapter.
package mage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/adapters"
	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/agentbase"
	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/authprobe"
	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/binaryutil"
	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/hookutil"
	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"

	_ "modernc.org/sqlite" // register sqlite driver for mage session metadata probes
)

const (
	// adapterID is the registry id and the value users pass to
	// `ao spawn --agent`. It matches domain.HarnessMage.
	adapterID = "mage"

	// mageAgentSessionIDMetadataKey is the session-metadata key the mage
	// plugin persists the native session id under. GetRestoreCommand reads it
	// back to resume an existing session.
	mageAgentSessionIDMetadataKey = "agentSessionId"

	// mageConfigEnvVar points mage at an AO-owned config file. mage honors it
	// exactly like opencode honors OPENCODE_CONFIG (verified against
	// `mage debug config`): the file merges over the user's global, custom,
	// and project configuration.
	mageConfigEnvVar = "MAGE_CONFIG"
)

var mageUnixPaths = []string{
	"/usr/local/bin/mage",
	"/opt/homebrew/bin/mage",
}

// Plugin is the mage agent adapter. It is safe for concurrent use; the binary
// path is resolved once and cached under binaryMu.
type Plugin struct {
	agentbase.Base
	binaryMu       sync.Mutex
	resolvedBinary string
}

// New returns a ready-to-register mage adapter.
func New() *Plugin {
	return &Plugin{}
}

// EmitsSemanticMessageAcceptance reports that the mage activity plugin
// includes the accepted prompt text in its user-prompt-submit callback.
func (p *Plugin) EmitsSemanticMessageAcceptance() bool { return true }

var _ adapters.Adapter = (*Plugin)(nil)
var _ ports.Agent = (*Plugin)(nil)
var _ ports.AgentAuthChecker = (*Plugin)(nil)
var _ ports.SemanticMessageAcceptanceSignaler = (*Plugin)(nil)

// Manifest returns the adapter's static self-description.
func (p *Plugin) Manifest() adapters.Manifest {
	return adapters.Manifest{
		ID:          adapterID,
		Name:        "mage",
		Description: "Run mage worker sessions.",
		Version:     "0.0.1",
		Capabilities: []adapters.Capability{
			adapters.CapabilityAgent,
		},
	}
}

// ResolveBinary returns the first mage binary on this machine, searching PATH
// then a handful of well-known install locations.
func (p *Plugin) ResolveBinary(ctx context.Context) (string, error) {
	p.binaryMu.Lock()
	defer p.binaryMu.Unlock()

	if p.resolvedBinary != "" {
		return p.resolvedBinary, nil
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	binary, err := resolveMageBinary(ctx)
	if err != nil {
		return "", err
	}
	p.resolvedBinary = binary
	return binary, nil
}

// GetConfigSpec reports mage's optional provider/model override.
func (p *Plugin) GetConfigSpec(ctx context.Context) (ports.ConfigSpec, error) {
	return agentbase.ModelConfigSpec(ctx, "Model override passed to `mage --model`.")
}

// GetLaunchCommand builds the argv to start a new interactive mage session.
// Shape:
//
//	[env MAGE_CONFIG=<ao-config>] mage [--auto] [--model <model>] [--agent <ao-agent>] [--prompt <prompt>]
//
// The session runs in the worktree (cwd is set by the runtime). mage has no
// CLI flag to set a system prompt, so AO writes a mage config into the AO
// prompt artifact directory, points MAGE_CONFIG at it, and selects the
// generated agent with --agent. The initial task prompt is delivered via
// --prompt (its argument, so a leading "-" is not read as a flag).
func (p *Plugin) GetLaunchCommand(ctx context.Context, cfg ports.LaunchConfig) (cmd []string, err error) {
	binary, err := p.ResolveBinary(ctx)
	if err != nil {
		return nil, err
	}

	envPrefix, agentName, err := mageConfigEnvPrefix(cfg.Permissions, cfg.SystemPrompt, cfg.SystemPromptFile, cfg.SessionID)
	if err != nil {
		return nil, err
	}
	cmd = envPrefix
	cmd = append(cmd, binary)
	appendPermissionFlags(&cmd, cfg.Permissions)
	agentbase.AppendModelFlag(&cmd, cfg.Config, "--model")
	if agentName != "" {
		cmd = append(cmd, "--agent", agentName)
	}
	if cfg.Prompt != "" {
		cmd = append(cmd, "--prompt", cfg.Prompt)
	}
	return cmd, nil
}

// GetRestoreCommand rebuilds the argv that continues an existing mage
// session: `[env MAGE_CONFIG=<ao-config>] mage [--auto] [--agent <ao-agent>] --session <agentSessionId> [--prompt <prompt>]`.
// It re-applies the permission flag and the generated AO agent config (resume
// otherwise reverts to configured defaults). ok is false when the
// plugin-derived native session id has not landed yet, so callers fall back
// to fresh launch behavior — mirroring the opencode adapter.
func (p *Plugin) GetRestoreCommand(ctx context.Context, cfg ports.RestoreConfig) (cmd []string, ok bool, err error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	agentSessionID := strings.TrimSpace(cfg.Session.Metadata[mageAgentSessionIDMetadataKey])
	if agentSessionID == "" {
		return nil, false, nil
	}

	binary, err := p.ResolveBinary(ctx)
	if err != nil {
		return nil, false, err
	}

	envPrefix, agentName, err := mageConfigEnvPrefix(cfg.Permissions, cfg.SystemPrompt, cfg.SystemPromptFile, cfg.Session.ID)
	if err != nil {
		return nil, false, err
	}
	cmd = envPrefix
	cmd = append(cmd, binary)
	appendPermissionFlags(&cmd, cfg.Permissions)
	agentbase.AppendModelFlag(&cmd, cfg.Config, "--model")
	if agentName != "" {
		cmd = append(cmd, "--agent", agentName)
	}
	cmd = append(cmd, "--session", agentSessionID)
	if cfg.Prompt != "" {
		cmd = append(cmd, "--prompt", cfg.Prompt)
	}
	return cmd, true, nil
}

// SessionInfo surfaces mage plugin-derived metadata. Metadata is
// intentionally nil for mage: callers get the normalized fields directly,
// matching the opencode adapter.
func (p *Plugin) SessionInfo(ctx context.Context, session ports.SessionRef) (ports.SessionInfo, bool, error) {
	if err := ctx.Err(); err != nil {
		return ports.SessionInfo{}, false, err
	}
	info, ok := agentbase.StandardSessionInfo(session)
	return info, ok, nil
}

// appendPermissionFlags uses mage's own approval flags where it has them.
// --auto approves whatever is not explicitly denied, which is exactly AO's
// auto mode, so the Terminal UI runs the provider's implementation rather
// than an emulation of it. Bypass carries its full-access rule on the AO
// agent, which outranks every config layer (mage has no
// --dangerously-skip-permissions alias).
func appendPermissionFlags(cmd *[]string, permissions ports.PermissionMode) {
	if ports.NormalizePermissionMode(permissions) == ports.PermissionModeAuto {
		*cmd = append(*cmd, "--auto")
	}
}

type mageInlineConfig struct {
	Schema     string                       `json:"$schema,omitempty"`
	Permission map[string]string            `json:"permission,omitempty"`
	Agent      map[string]mageAgentSettings `json:"agent,omitempty"`
}

type mageAgentSettings struct {
	Mode   string `json:"mode,omitempty"`
	Prompt string `json:"prompt,omitempty"`
	// Either a ruleset or mage's scalar form, and it outranks every config
	// layer — which is what makes bypass a true bypass.
	Permission any `json:"permission,omitempty"`
}

// magePermissionConfig maps AO's approval vocabulary onto what mage itself
// does, so a mode behaves the way the same session would outside AO.
//
//   - default: no rule. mage's own configuration decides, unchanged.
//   - accept-edits: allow edits. mage has no such mode, so this is the one
//     grant AO adds; everything else stays as mage would have it.
//   - auto: mage's --auto, appended at launch, which approves whatever is
//     not explicitly denied.
//   - bypass-permissions: `permission: "allow"` on the AO agent, mage's own
//     full-access setting, which outranks every config layer.
func magePermissionConfig(mode ports.PermissionMode) map[string]string {
	if ports.NormalizePermissionMode(mode) == ports.PermissionModeAcceptEdits {
		return map[string]string{"edit": "allow"}
	}
	return nil
}

func mageConfigEnvPrefix(permissions ports.PermissionMode, inlinePrompt, promptFile, sessionID string) ([]string, string, error) {
	permission := magePermissionConfig(permissions)
	// Bypass goes on the agent, not in the config body: agent rules outrank
	// every config layer, so the mode matches what Chat does.
	var agentPermission any
	if ports.NormalizePermissionMode(permissions) == ports.PermissionModeBypassPermissions {
		agentPermission, permission = "allow", nil
	}
	// Both overlays ride on the AO-owned config file, never on
	// MAGE_CONFIG_CONTENT: that variable is the highest-precedence config
	// source and callers own it. Writing to MAGE_CONFIG also keeps a single
	// answer for precedence: a worktree mage.json always overrides AO.
	if promptFile == "" {
		if inlinePrompt != "" {
			return nil, "", fmt.Errorf("mage: system prompt file required to build agent config")
		}
		return nil, "", nil
	}
	dir := filepath.Dir(promptFile)
	configPath := filepath.Join(dir, "mage.json")
	agentName := mageAOAgentName(sessionID)
	prompt := inlinePrompt
	if prompt == "" {
		prompt = "{file:./" + filepath.Base(promptFile) + "}"
	}
	config := mageInlineConfig{
		Schema:     "https://opencode.ai/config.json",
		Permission: permission,
		Agent: map[string]mageAgentSettings{
			agentName: {Mode: "primary", Prompt: prompt, Permission: agentPermission},
		},
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, "", err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, "", fmt.Errorf("mage: create prompt config dir: %w", err)
	}
	if err := hookutil.AtomicWriteFile(configPath, data, 0o600); err != nil {
		return nil, "", fmt.Errorf("mage: write prompt config: %w", err)
	}
	return []string{"env", mageConfigEnvVar + "=" + configPath}, agentName, nil
}

// PrepareACPConfigContent merges AO's standing instructions and its approval
// tiers into mage's inline runtime overlay (MAGE_CONFIG_CONTENT). The user's
// own mage configuration remains untouched.
func PrepareACPConfigContent(
	existing, systemPrompt string,
	permissions ports.PermissionMode,
) (string, error) {
	content, err := prepareACPConfigContent(existing, systemPrompt, permissions)
	if err != nil {
		return "", fmt.Errorf("mage: %w", err)
	}
	return content, nil
}

func mageAOAgentName(sessionID string) string {
	const fallback = "ao-system-prompt"
	trimmed := strings.TrimSpace(sessionID)
	if trimmed == "" {
		return fallback
	}
	var b strings.Builder
	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-',
			r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	name := strings.Trim(b.String(), "-_")
	if name == "" {
		return fallback
	}
	return "ao-" + name
}

// AuthStatus checks whether mage has a configured provider credential.
// Missing credentials remain unknown because mage can still run public free
// models without a provider login. mage's own stored credentials decide
// first, then `mage auth list`, and only then a key in the environment.
func (p *Plugin) AuthStatus(ctx context.Context) (ports.AgentAuthStatus, error) {
	binary, err := p.ResolveBinary(ctx)
	if err != nil {
		return ports.AgentAuthStatusUnknown, err
	}
	status, ok, storedErr := mageStoredAuthStatus(ctx)
	if storedErr == nil && ok && status.HasCredential() {
		return status, nil
	}
	if storedErr == nil {
		listed, err := mageAuthListStatus(ctx, binary)
		if err != nil {
			return ports.AgentAuthStatusUnknown, err
		}
		if listed == ports.AgentAuthStatusAuthorized {
			return listed, nil
		}
	}
	if envStatus, envOK := mageEnvAuthStatus(ctx); envOK {
		return envStatus, nil
	}
	if storedErr != nil {
		return ports.AgentAuthStatusUnknown, storedErr
	}
	if ok {
		return status, nil
	}
	return ports.AgentAuthStatusUnknown, nil
}

func mageAuthListStatus(ctx context.Context, binary string) (status ports.AgentAuthStatus, resultErr error) {
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	started := time.Now()
	out, err := authprobe.CmdRunner(probeCtx, binary, "auth", "list")
	noCredentials := false
	defer func() {
		step := authprobe.CommandStep(binary, []string{"auth", "list"}, status, err, probeCtx.Err(), started)
		if noCredentials && probeCtx.Err() == nil {
			source := step.Source
			step = domain.NewAgentAuthDiagnostic(step.Method, "credential_missing")
			step.Source = source
		}
		ports.RecordAuthDiagnostic(ctx, step)
	}()
	if probeCtx.Err() != nil {
		if probeCtx.Err() == context.DeadlineExceeded && ctx.Err() == nil {
			return ports.AgentAuthStatusUnknown, nil
		}
		return ports.AgentAuthStatusUnknown, probeCtx.Err()
	}
	text := strings.ToLower(string(out))
	if strings.Contains(text, "0 credentials") || strings.Contains(text, "no credentials") || strings.Contains(text, "not authenticated") {
		noCredentials = true
		return ports.AgentAuthStatusUnknown, nil
	}
	if strings.Contains(text, "credential") && err == nil {
		ports.RecordAuthEvidence(ctx, domain.CommandSource(binary, "auth", "list"))
		return ports.AgentAuthStatusAuthorized, nil
	}
	return ports.AgentAuthStatusUnknown, nil
}

var mageAPIKeyEnvVars = []string{
	"MAGE_API_KEY",
	"OPENAI_API_KEY",
	"ANTHROPIC_API_KEY",
	"GEMINI_API_KEY",
	"GOOGLE_API_KEY",
	"OPENROUTER_API_KEY",
	"DEEPSEEK_API_KEY",
	"GROQ_API_KEY",
	"XAI_API_KEY",
	"MISTRAL_API_KEY",
	"COHERE_API_KEY",
}

func mageEnvAuthStatus(ctx context.Context) (ports.AgentAuthStatus, bool) {
	if authprobe.CredentialEnvironment(ctx, domain.AgentAccountMethodAPIKey, mageAPIKeyEnvVars...) != "" {
		return ports.AgentAuthStatusAuthorized, true
	}
	return ports.AgentAuthStatusUnknown, false
}

// mageStoredAuthStatus reads mage's auth.json, then its database. mage keeps
// both under $XDG_DATA_HOME/mage (default ~/.local/share/mage), mirroring
// opencode's layout with its own directory and db filename.
func mageStoredAuthStatus(ctx context.Context) (ports.AgentAuthStatus, bool, error) {
	if err := ctx.Err(); err != nil {
		return ports.AgentAuthStatusUnknown, false, err
	}
	dataDir, ok := mageDataDir()
	if !ok {
		return ports.AgentAuthStatusUnknown, false, nil
	}
	authPath := filepath.Join(dataDir, "auth.json")
	jsonStatus, jsonOK, data, err := mageAuthJSONStatus(ctx, authPath)
	if err != nil {
		return ports.AgentAuthStatusUnknown, false, err
	}
	if jsonOK && jsonStatus == ports.AgentAuthStatusAuthorized {
		evidence := authprobe.FileSource(authPath).WithMethod(authprobe.JSONProviderMethod(data))
		evidence.Provider = authprobe.JSONProviderNames(data)
		ports.RecordAuthEvidence(ctx, evidence)
		return jsonStatus, true, nil
	}
	dbPath := filepath.Join(dataDir, "mage.db")
	if status, ok, err := mageDBAuthStatus(ctx, dbPath); err != nil || ok {
		if status == ports.AgentAuthStatusAuthorized {
			ports.RecordAuthEvidence(ctx, authprobe.FileSource(dbPath).WithMethod(domain.AgentAccountMethodAccount))
		}
		return status, ok, err
	}
	if jsonOK {
		return jsonStatus, true, nil
	}
	return ports.AgentAuthStatusUnknown, false, nil
}

func mageDataDir() (string, bool) {
	if dataDir := strings.TrimSpace(os.Getenv("MAGE_DATA_DIR")); dataDir != "" {
		return dataDir, true
	}
	if dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); dataHome != "" {
		return filepath.Join(dataHome, "mage"), true
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", false
	}
	return filepath.Join(home, ".local", "share", "mage"), true
}

// mageAuthJSONStatus also returns the file's contents, so callers parse it
// without reading it again.
func mageAuthJSONStatus(ctx context.Context, path string) (ports.AgentAuthStatus, bool, []byte, error) {
	data, err := authprobe.ReadCredentialFile(ctx, path)
	if os.IsNotExist(err) {
		return ports.AgentAuthStatusUnknown, false, nil, nil
	}
	if err != nil {
		return ports.AgentAuthStatusUnknown, false, nil, err
	}
	if strings.TrimSpace(string(data)) == "" {
		return ports.AgentAuthStatusUnknown, true, data, nil
	}

	var entries map[string]json.RawMessage
	if err := json.Unmarshal(data, &entries); err != nil {
		authprobe.RecordFileDiagnostic(ctx, path, "credential_file_invalid")
		return ports.AgentAuthStatusUnknown, false, nil, err
	}
	if len(entries) == 0 {
		return ports.AgentAuthStatusUnknown, true, data, nil
	}
	for key, value := range entries {
		if strings.TrimSpace(key) == "" {
			continue
		}
		trimmed := strings.TrimSpace(string(value))
		if trimmed != "" && trimmed != "null" && trimmed != "{}" {
			return ports.AgentAuthStatusAuthorized, true, data, nil
		}
	}
	return ports.AgentAuthStatusUnknown, true, data, nil
}

func mageDBAuthStatus(ctx context.Context, path string) (status ports.AgentAuthStatus, ok bool, err error) {
	defer func() {
		if err != nil && ctx.Err() == nil {
			authprobe.RecordFileDiagnostic(ctx, path, "credential_store_unreadable")
		}
	}()
	if err := ctx.Err(); err != nil {
		return ports.AgentAuthStatusUnknown, false, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		authprobe.RecordFileDiagnostic(ctx, path, "credential_file_missing")
		return ports.AgentAuthStatusUnknown, false, nil
	} else if err != nil {
		return ports.AgentAuthStatusUnknown, false, err
	}

	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro&_pragma=busy_timeout(1000)")
	if err != nil {
		return ports.AgentAuthStatusUnknown, false, err
	}
	defer func() {
		_ = db.Close()
	}()

	authorized, known, err := mageDBHasAuthorizedAccount(ctx, db)
	if err != nil {
		return ports.AgentAuthStatusUnknown, false, err
	}
	if !known {
		return ports.AgentAuthStatusUnknown, false, nil
	}
	if authorized {
		return ports.AgentAuthStatusAuthorized, true, nil
	}
	return ports.AgentAuthStatusUnknown, true, nil
}

func mageDBHasAuthorizedAccount(ctx context.Context, db *sql.DB) (authorized, known bool, err error) {
	for _, query := range []string{
		`SELECT COUNT(*) FROM credential WHERE active = 1 AND trim(value) != ''`,
		`SELECT COUNT(*) FROM account_state WHERE active_account_id IS NOT NULL AND trim(active_account_id) != ''`,
		`SELECT COUNT(*) FROM account WHERE trim(access_token) != ''`,
		`SELECT COUNT(*) FROM control_account WHERE active = 1 AND trim(access_token) != ''`,
	} {
		count, err := mageDBCount(ctx, db, query)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "no such table") {
				continue
			}
			return false, false, err
		}
		known = true
		if count > 0 {
			return true, true, nil
		}
	}
	return false, known, nil
}

func mageDBCount(ctx context.Context, db *sql.DB, query string) (int, error) {
	var count int
	if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// resolveMageBinary returns the first mage executable in resolution order:
// the LookPath winner, other PATH entries, then well-known install locations.
func resolveMageBinary(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	names := []string{"mage"}
	if runtime.GOOS == "windows" {
		names = []string{"mage.cmd", "mage.exe", "mage"}
	}
	if path, err := exec.LookPath(names[0]); err == nil {
		return path, nil
	}

	var candidates []string
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			candidates = append(candidates,
				filepath.Join(appData, "npm", "mage.cmd"),
				filepath.Join(appData, "npm", "mage.exe"),
			)
		}
		candidates = append(candidates, binaryutil.WindowsPackageManagerBinCandidates("mage")...)
	} else {
		candidates = append(candidates, mageUnixPaths...)
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates,
				filepath.Join(home, ".npm-global", "bin", "mage"),
				filepath.Join(home, ".npm", "bin", "mage"),
				filepath.Join(home, ".local", "bin", "mage"),
			)
			candidates = append(candidates, binaryutil.UnixPackageManagerBinCandidates(home, "mage")...)
			nodeManagerCandidates, err := binaryutil.UnixNodeManagerBinCandidates(ctx, home, "mage")
			if err != nil {
				return "", err
			}
			candidates = append(candidates, nodeManagerCandidates...)
		}
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		for _, name := range names {
			candidates = append(candidates, filepath.Join(dir, name))
		}
	}
	for _, candidate := range candidates {
		if hookutil.IsExecutableFile(candidate) {
			return candidate, nil
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("mage: %w", ports.ErrAgentBinaryNotFound)
}
