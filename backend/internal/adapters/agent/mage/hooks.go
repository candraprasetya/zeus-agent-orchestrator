package mage

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/hookutil"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
	"github.com/aoagents/agent-orchestrator/backend/internal/skillassets"
)

const (
	// mage scans both `.mage/plugin/` and `.mage/plugins/` for `*.js`/`*.ts`
	// files (its ConfigPlugin glob "{plugin,plugins}/*.{ts,js}" is inherited
	// from the opencode config schema mage implements). AO writes the plural
	// `plugins/`, matching the directory opencode's tooling uses.
	magePluginDirName = ".mage"
	magePluginSubDir  = "plugins"

	// magePluginFileName is the AO-owned plugin file. AO fully owns this
	// filename: install overwrites it and uninstall deletes it (guarded by the
	// sentinel), so user-authored plugins in other files are never touched.
	// It is TypeScript (mage runs on Bun). Unlike the opencode plugin it has
	// no imports at all — not even type-only ones — because mage's plugin
	// loader resolves import specifiers eagerly, and a type import of an
	// uninstalled package stalls plugin boot.
	magePluginFileName = "ao-activity.ts"

	// magePluginSentinel marks the file as AO-managed. AreHooksInstalled and
	// UninstallHooks key off it so AO never deletes a user file that happens to
	// share the name. It must appear verbatim in the embedded plugin source.
	magePluginSentinel = "agent-orchestrator: managed mage activity plugin"

	// mageHookCommandPrefix identifies the hook commands AO owns. The embedded
	// plugin shells `ao hooks mage <event>`; this prefix is the shared contract
	// with the `ao hooks` CLI and is asserted by tests so the plugin can't
	// silently drift away from it.
	mageHookCommandPrefix = "ao hooks mage "

	// mageSkillSubDir is where mage discovers project skills
	// (`.mage/skills/<name>/SKILL.md`). AO materializes the using-ao skill
	// here so mage's native `skill` tool can see it — the data-dir install
	// alone is invisible to that discovery path.
	mageSkillSubDir = "skills"

	// mageSkillMarkerFile lives beside the skill directory (not inside it) so
	// Materialize's RemoveAll of using-ao/ cannot erase ownership mid-install.
	// Install overwrites and uninstall deletes only when this marker is present.
	mageSkillMarkerFile = ".using-ao.ao-managed"

	// mageSkillSentinel is written into the marker file. Keep it distinct
	// from the plugin sentinel so ownership checks stay file-specific.
	mageSkillSentinel = "agent-orchestrator: managed mage using-ao skill"
)

// magePluginSource is the AO-managed mage plugin, embedded so it ships inside
// the binary and is written verbatim into a session's worktree on hook install.
//
//go:embed assets/ao-activity.ts
var magePluginSource string

// mageManagedEvents are the normalized activity events the embedded plugin
// reports. They are defined here (not parsed from the file) so tests can assert
// the plugin wires every one via the `ao hooks mage <event>` command.
var mageManagedEvents = []string{"session-start", "user-prompt-submit", "active", "stop", "permission-blocked"}

// GetAgentHooks installs AO's mage activity plugin into the worktree-local
// .mage/plugins/ directory, and materializes the using-ao skill into
// .mage/skills/using-ao/ so mage's native `skill` tool can discover it. mage
// has no native command-hook config to merge into; its only
// lifecycle-extensibility surface is a JS/TS plugin, so AO writes a dedicated,
// AO-owned plugin file. The write is atomic and idempotent: re-installing
// overwrites AO's own file with identical content. It refuses to overwrite a
// file that is NOT AO-managed (no sentinel), so a user plugin that happens to
// occupy our path is never silently destroyed — install fails loudly instead.
// The skill install uses the same ownership guard via a marker file beside the
// skill directory (written before Materialize runs).
func (p *Plugin) GetAgentHooks(ctx context.Context, cfg ports.WorkspaceHookConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.WorkspacePath) == "" {
		return errors.New("mage.GetAgentHooks: WorkspacePath is required")
	}

	pluginPath := magePluginPath(cfg.WorkspacePath)
	// Guard against clobbering a user file at our path: overwrite only when the
	// target is absent or already AO-managed. A foreign file is a loud error,
	// not silent data loss (uninstall is sentinel-guarded the same way).
	if _, err := os.Stat(pluginPath); err == nil {
		managed, err := isAOManagedPlugin(pluginPath)
		if err != nil {
			return fmt.Errorf("mage.GetAgentHooks: %w", err)
		}
		if !managed {
			return fmt.Errorf("mage.GetAgentHooks: refusing to overwrite non-AO file at %s — move it so AO can install its plugin", pluginPath)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("mage.GetAgentHooks: stat plugin: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(pluginPath), 0o750); err != nil {
		return fmt.Errorf("mage.GetAgentHooks: create plugin dir: %w", err)
	}
	if err := hookutil.AtomicWriteFile(pluginPath, []byte(magePluginSource), 0o600); err != nil {
		return fmt.Errorf("mage.GetAgentHooks: write plugin: %w", err)
	}
	if err := hookutil.EnsureWorkspaceGitignore(filepath.Dir(pluginPath), magePluginFileName); err != nil {
		return fmt.Errorf("mage.GetAgentHooks: gitignore: %w", err)
	}
	if err := InstallUsingAOSkill(cfg.WorkspacePath); err != nil {
		return fmt.Errorf("mage.GetAgentHooks: %w", err)
	}
	return nil
}

// UninstallHooks removes AO's mage plugin and the AO-managed using-ao skill
// from the workspace-local .mage/ tree. It deletes the plugin only when it
// carries the AO sentinel, and the skill directory only when the AO marker is
// present, so user files that happen to share those paths are left in place. A
// missing file is a no-op.
func (p *Plugin) UninstallHooks(ctx context.Context, workspacePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(workspacePath) == "" {
		return errors.New("mage.UninstallHooks: workspacePath is required")
	}

	pluginPath := magePluginPath(workspacePath)
	managed, err := isAOManagedPlugin(pluginPath)
	if err != nil {
		return fmt.Errorf("mage.UninstallHooks: %w", err)
	}
	if managed {
		if err := os.Remove(pluginPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("mage.UninstallHooks: remove plugin: %w", err)
		}
	}
	if err := UninstallUsingAOSkill(workspacePath); err != nil {
		return fmt.Errorf("mage.UninstallHooks: %w", err)
	}
	return nil
}

// AreHooksInstalled reports whether AO's mage plugin is present in the
// workspace-local plugin dir. A missing file, or a same-named file without the
// AO sentinel, means none are installed.
func (p *Plugin) AreHooksInstalled(ctx context.Context, workspacePath string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if strings.TrimSpace(workspacePath) == "" {
		return false, errors.New("mage.AreHooksInstalled: workspacePath is required")
	}
	managed, err := isAOManagedPlugin(magePluginPath(workspacePath))
	if err != nil {
		return false, fmt.Errorf("mage.AreHooksInstalled: %w", err)
	}
	return managed, nil
}

func magePluginPath(workspacePath string) string {
	return filepath.Join(workspacePath, magePluginDirName, magePluginSubDir, magePluginFileName)
}

func mageSkillDir(workspacePath string) string {
	return filepath.Join(workspacePath, magePluginDirName, mageSkillSubDir, skillassets.SkillName)
}

func mageSkillsDir(workspacePath string) string {
	return filepath.Join(workspacePath, magePluginDirName, mageSkillSubDir)
}

func mageSkillMarkerPath(workspacePath string) string {
	return filepath.Join(mageSkillsDir(workspacePath), mageSkillMarkerFile)
}

// InstallUsingAOSkill materializes the embedded shared using-ao skill into
// .mage/skills/using-ao/. It refuses to overwrite a same-named directory that
// is not AO-managed.
func InstallUsingAOSkill(workspacePath string) error {
	skillDir := mageSkillDir(workspacePath)
	if info, err := os.Stat(skillDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("refusing to overwrite non-directory at %s — move it so AO can install using-ao", skillDir)
		}
		managed, err := isAOManagedSkill(workspacePath)
		if err != nil {
			return err
		}
		if !managed {
			return fmt.Errorf("refusing to overwrite non-AO skill at %s — move it so AO can install using-ao", skillDir)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat skill dir: %w", err)
	}

	skillsParent := mageSkillsDir(workspacePath)
	if err := os.MkdirAll(skillsParent, 0o750); err != nil {
		return fmt.Errorf("create skills dir: %w", err)
	}
	// Write ownership before Materialize clobbers using-ao/, so a crash mid-tree
	// write leaves a marker that allows the next install attempt to recover.
	if err := hookutil.AtomicWriteFile(mageSkillMarkerPath(workspacePath), []byte(mageSkillSentinel+"\n"), 0o600); err != nil {
		return fmt.Errorf("write skill marker: %w", err)
	}
	if err := skillassets.Materialize(skillDir); err != nil {
		return fmt.Errorf("materialize using-ao skill: %w", err)
	}
	if err := ensureSkillTreeGitignored(skillDir); err != nil {
		return fmt.Errorf("skill gitignore: %w", err)
	}
	if err := hookutil.EnsureWorkspaceGitignore(skillsParent, mageSkillMarkerFile); err != nil {
		return fmt.Errorf("skill marker gitignore: %w", err)
	}
	return nil
}

// ensureSkillTreeGitignored writes AO-managed .gitignore files beside every
// file in the skill tree so registry's hook-footprint contract holds: each
// installed path must be ignorable for git worktree teardown.
func ensureSkillTreeGitignored(skillRoot string) error {
	byDir := map[string][]string{}
	err := filepath.WalkDir(skillRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		dir := filepath.Dir(path)
		byDir[dir] = append(byDir[dir], filepath.Base(path))
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk skill tree: %w", err)
	}
	for dir, names := range byDir {
		if err := hookutil.EnsureWorkspaceGitignore(dir, names...); err != nil {
			return err
		}
	}
	return nil
}

// UninstallUsingAOSkill removes the shared skill only when its AO marker owns
// it. A missing directory or marker is a no-op.
func UninstallUsingAOSkill(workspacePath string) error {
	managed, err := isAOManagedSkill(workspacePath)
	if err != nil {
		return err
	}
	if !managed {
		return nil
	}
	if err := os.RemoveAll(mageSkillDir(workspacePath)); err != nil {
		return fmt.Errorf("remove skill dir: %w", err)
	}
	markerPath := mageSkillMarkerPath(workspacePath)
	if err := os.Remove(markerPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove skill marker: %w", err)
	}
	return nil
}

// isAOManagedPlugin reports whether the file at path exists and carries the AO
// sentinel. A missing file yields (false, nil).
func isAOManagedPlugin(path string) (bool, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path built from caller-owned workspace dir
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	return strings.Contains(string(data), magePluginSentinel), nil
}

// isAOManagedSkill reports whether the AO ownership marker beside the skill
// directory exists. A missing marker yields (false, nil).
func isAOManagedSkill(workspacePath string) (bool, error) {
	data, err := os.ReadFile(mageSkillMarkerPath(workspacePath)) //nolint:gosec // path built from caller-owned workspace dir
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read skill marker: %w", err)
	}
	return strings.Contains(string(data), mageSkillSentinel), nil
}
