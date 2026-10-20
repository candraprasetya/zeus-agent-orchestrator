# Mage Adapter

Mage is integrated into Agent Orchestrator as an interactive Terminal UI
harness with Chat over its native ACP server. AO launches the user's own
`mage` binary inside the session worktree. Mage implements the
`https://opencode.ai/config.json` config schema and the same
`{plugin,plugins}/*.{ts,js}` plugin glob as OpenCode, so AO bridges it with
the same mechanisms: an AO-owned config overlay for standing instructions and
an AO-owned workspace plugin for activity.

## Install

Install mage with npm (its only documented method), then make sure `mage` is
available on `PATH`:

```bash
npm install -g @mybcabisnis/mage
```

AO resolves `mage` from:

- `PATH`
- `/usr/local/bin/mage`, `/opt/homebrew/bin/mage`
- `~/.npm-global/bin/mage`, `~/.npm/bin/mage`, `~/.local/bin/mage`
- the usual npm/pnpm/yarn/bun global prefixes

## Setup

Mage stores its credentials under `$XDG_DATA_HOME/mage` (default
`~/.local/share/mage`), including `auth.json` and `mage.db`. Until a provider
credential exists the outcome stays "unknown" rather than unauthenticated,
because Mage can still run free models without a login. Harness settings offer
**Log in to Mage**, which opens `mage auth login` (the native provider
chooser) in an AO terminal.

When the agent config sets a model, AO appends `--model <provider/model>`
(verified against `mage models`, which prints `provider/model` ids such as
`merlin/default`).

## Supported AO Modes

Mage is exposed through AO's Terminal UI mode and through Chat (native ACP).
A fresh prompted session launches as:

```bash
[env MAGE_CONFIG=<ao-config>] mage [--auto] [--model <model>] [--agent <ao-agent>] [--prompt <prompt>]
```

Mage has no system-prompt flag, so AO writes a Mage config into the AO prompt
artifact directory, points `MAGE_CONFIG` at it, and selects the generated
agent with `--agent`. A restored session resumes the native conversation
reported by the activity plugin:

```bash
[env MAGE_CONFIG=<ao-config>] mage [--auto] [--agent <ao-agent>] --session <agentSessionId> [--prompt <prompt>]
```

AO permission modes map onto Mage's own policy:

| AO mode | Mage |
| --- | --- |
| default | no rule: the user's own Mage configuration decides |
| accept-edits | config `permission: {edit: "allow"}` (Mage has no edits-only mode) |
| auto | `--auto`: auto-approves whatever is not explicitly denied |
| bypass | `permission: "allow"` on the AO agent, which outranks every config layer |

Mage has no `--dangerously-skip-permissions` alias; bypass therefore rides on
the agent config rather than a launch flag.

## Instructions

Mage has no system-prompt flag. AO delivers its standing instructions through
the generated agent in the `MAGE_CONFIG` overlay, either inlined or as a
`{file:...}` reference to the prompt artifact AO already writes per session.

## Activity Tracking

AO installs an AO-owned plugin at `.mage/plugins/ao-activity.ts` inside the
session worktree, plus the using-ao skill under `.mage/skills/using-ao/`. The
plugin normalizes Mage's native events onto `ao hooks mage <event>`
callbacks:

- `session.created` reports `active` and the native session id
- `message.updated` / `message.part.updated` report `active` and the accepted prompt text
- `tool.execute.before` / `tool.execute.after` report `active`
- `session.status` (idle) reports `idle`
- `permission.asked` / `question.asked` report `blocked`

AO fully owns the plugin filename: install overwrites only files carrying the
AO sentinel and refuses to touch a user-authored file at the same path.
Hook delivery is best-effort; a missing `ao` executable or a hook timeout
never interrupts the Mage session.

## Chat Mode

Chat launches `mage acp` and feeds AO's standing instructions and approval
tier through the `MAGE_CONFIG_CONTENT` overlay, the same precedence-free
mechanism OpenCode's Chat driver uses. Model overrides must use
`provider/model` form as printed by `mage models`.

## Current Limits

- The plugin-derived native session id is required for exact restore. If the
  plugin never reports (for example a Mage build that does not load project
  plugins), a restore falls back to a fresh launch.
- Mage is not a reviewer harness.
