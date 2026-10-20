// agent-orchestrator: managed mage activity plugin (do not edit)
//
// It maps mage's native lifecycle events onto AO's normalized activity
// events, mirroring the OpenCode activity plugin because mage is an
// opencode-compatible fork (same config schema, same plugin loader):
//   session.created                       -> `ao hooks mage session-start`
//   message.updated / message.part.updated -> `ao hooks mage user-prompt-submit`
//   tool.execute.before / tool.execute.after -> `ao hooks mage active`
//   session.status (status.type == idle)   -> `ao hooks mage stop`
//   permission.asked / question.asked      -> `ao hooks mage permission-blocked`
//   permission/question reply or rejection -> `ao hooks mage active`
//
// The mage-native session id (and prompt/model where known) is piped to the
// hook command as JSON on stdin, run with cwd set to the worktree so AO can
// correlate the mage session to its AO session. Every invocation is
// best-effort and must never crash the user's mage session: a missing `ao`
// binary is a guarded no-op (`Bun.which`), and spawn exceptions, non-zero
// exit codes, and malformed event payloads are caught and surfaced through
// mage's structured logger (client.app.log) for diagnosis — never rethrown.
//
// This file deliberately has no imports (not even a type-only one): mage's
// plugin loader resolves imports eagerly, and a type-only import of an
// uninstalled package stalls plugin boot. All types are structural.

const HOOK_TIMEOUT_MS = 2000

export default async function aoActivity({ directory, client }: any) {
  // Dedup state: message.* events don't carry the session id, so track it
  // from events that do. Maps message id to whether the report we already
  // sent included the prompt text.
  const promptReports = new Map<string, boolean>()
  // message.* events don't carry the session id, so track it from events that do.
  let currentSessionID: string | null = null
  // The model of the most recent assistant message, forwarded for context.
  let currentModel: string | null = null
  // AO fences each provider generation with AO_RUNTIME_LAUNCH_ID; carry it in
  // the payload so hooks work even when child-process env inheritance is
  // trimmed by the host runtime.
  const launchID: string = (process.env.AO_RUNTIME_LAUNCH_ID ?? "").trim()
  const messageStore = new Map<string, any>()

  // Resolve AO through Bun so the same command works on Unix and Windows. A
  // missing binary remains a silent no-op rather than a per-event error.
  function hookCmd(hookName: string): string[] | null {
    const ao = Bun.which("ao")
    return ao ? [ao, "hooks", "mage", hookName] : null
  }

  // Report a hook failure through mage's structured logger. Best-effort: the
  // log call must itself never throw or reject back into mage, hence the
  // optional chaining + swallowed rejection.
  function logHookFailure(hookName: string, detail: string) {
    try {
      void client?.app
        ?.log?.({ body: { service: "ao-activity", level: "error", message: `hook ${hookName} failed: ${detail}` } })
        ?.catch?.(() => {})
    } catch {
      // The logger itself is unavailable — nothing more we can safely do.
    }
  }

  // All hooks are dispatched synchronously (Bun.spawnSync), for two reasons:
  //   1. Ordering. An async hook yields the event loop; if mage does not
  //      await the handler's promise, a later event (e.g. message.updated ->
  //      user-prompt-submit) could complete before an in-flight async
  //      session-start, so AO would see the prompt before the session is
  //      registered. spawnSync blocks mage's single-threaded loop until the
  //      hook returns, so events are reported strictly in dispatch order.
  //   2. `mage run` exits on the idle event, so an async stop hook would be
  //      killed before completing.
  //
  // A non-zero exit (the guard makes a missing `ao` exit 0, so this is a real
  // `ao hooks` failure) or a spawn exception is logged with its stderr and never
  // rethrown, so reporting failures are diagnosable without crashing mage.
  function callHookSync(hookName: string, payload: Record<string, unknown>) {
    try {
      const command = hookCmd(hookName)
      if (!command) return
      const result = Bun.spawnSync(command, {
        cwd: directory,
        env: { ...process.env, AO_RUNTIME_LAUNCH_ID: launchID },
        stdin: new TextEncoder().encode(JSON.stringify({ ...payload, launch_id: launchID }) + "\n"),
        stdout: "ignore",
        stderr: "pipe",
        timeout: HOOK_TIMEOUT_MS,
      })
      if (!result.success) {
        const stderr = result.stderr ? new TextDecoder().decode(result.stderr).trim() : ""
        logHookFailure(hookName, `exited ${result.exitCode}${stderr ? `: ${stderr}` : ""}`)
      }
    } catch (err) {
      // The spawn itself failed. Never propagate.
      logHookFailure(hookName, err instanceof Error ? err.message : String(err))
    }
  }

  function switchedSession(sessionID: string): boolean {
    if (currentSessionID === sessionID) return false
    promptReports.clear()
    messageStore.clear()
    currentModel = null
    currentSessionID = sessionID
    return true
  }

  // Report a user prompt, preferring the one that carries the prompt text.
  // message.updated can arrive before message.part.updated with no text, so an
  // early empty report must NOT dedup away the later text report — otherwise
  // the prompt never reaches AO and title-from-prompt metadata breaks. Therefore:
  // an empty report fires at most once (so run-mode flows that omit the text
  // part still mark the session active), and a text report fires once and is
  // terminal.
  function reportUserPrompt(sessionID: string, messageID: string, prompt: string) {
    const hasText = prompt.length > 0
    const reportedWithText = promptReports.get(messageID)
    if (reportedWithText) return // already reported with text — terminal
    if (reportedWithText === false && !hasText) return // already reported empty; no new info
    promptReports.set(messageID, hasText)
    callHookSync("user-prompt-submit", { session_id: sessionID, prompt, model: currentModel ?? "" })
  }

  return {
    event: async ({ event }: any) => {
      try {
        switch (event.type) {
          case "session.created": {
            const session = event.properties?.info
            if (!session?.id) break
            if (switchedSession(session.id)) {
              callHookSync("session-start", { session_id: session.id })
            }
            break
          }

          case "message.updated": {
            const msg = event.properties?.info
            if (!msg) break
            if (msg.sessionID && switchedSession(msg.sessionID)) {
              callHookSync("session-start", { session_id: msg.sessionID })
            }
            if (msg.role === "assistant" && msg.modelID) currentModel = msg.modelID
            // Fallback: some `mage run` flows never deliver message.part.updated
            // for the prompt, so start the turn from the user message itself.
            if (msg.role === "user") {
              messageStore.set(msg.id, msg)
              const sessionID = msg.sessionID ?? currentSessionID
              if (sessionID) reportUserPrompt(sessionID, msg.id, "")
            }
            break
          }

          case "message.part.updated": {
            const part = event.properties?.part
            if (!part?.messageID) break
            const msg = messageStore.get(part.messageID)
            if (msg?.role === "user" && part.type === "text") {
              const sessionID = msg.sessionID ?? currentSessionID
              const prompt = part.text ?? ""
              if (sessionID) reportUserPrompt(sessionID, msg.id, prompt)
              if (prompt.length > 0) messageStore.delete(part.messageID)
            }
            break
          }

          case "session.status": {
            // session.status fires in both TUI and `mage run`; session.idle is
            // deprecated and not reliably emitted in run mode.
            // AO's "stop" hook means "the current turn is idle/finished", not
            // "the whole native session has terminated", so multi-turn TUI
            // sessions intentionally emit one stop per idle transition.
            const props = event.properties
            if (props?.status?.type !== "idle") break
            const sessionID = props?.sessionID ?? currentSessionID
            if (!sessionID) break
            callHookSync("stop", { session_id: sessionID, model: currentModel ?? "" })
            break
          }

          case "permission.asked":
          case "question.asked": {
            const sessionID = event.properties?.sessionID ?? currentSessionID
            if (!sessionID) break
            callHookSync("permission-blocked", { session_id: sessionID, model: currentModel ?? "" })
            break
          }

          case "permission.updated":
          case "permission.replied":
          case "question.replied":
          case "permission.updated.answer":
          case "question.answer": {
            const sessionID = event.properties?.sessionID ?? currentSessionID
            if (!sessionID) break
            callHookSync("active", { session_id: sessionID, model: currentModel ?? "" })
            break
          }
        }
      } catch {
        // A malformed payload must never break the user's session.
      }
    },

    "tool.execute.before": async ({ tool }: any) => {
      try {
        const sessionID = tool?.sessionID ?? currentSessionID
        if (!sessionID) return
        callHookSync("active", { session_id: sessionID, model: currentModel ?? "" })
      } catch {
        // Best-effort only.
      }
    },

    "tool.execute.after": async ({ tool }: any) => {
      try {
        const sessionID = tool?.sessionID ?? currentSessionID
        if (!sessionID) return
        callHookSync("active", { session_id: sessionID, model: currentModel ?? "" })
      } catch {
        // Best-effort only.
      }
    },
  }
}
