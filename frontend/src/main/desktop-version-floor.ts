import { app, dialog, shell } from "electron";
import semver from "semver";

const FLOOR_URL =
  "https://raw.githubusercontent.com/Untrivial-ai/agent-orchestrator/main/desktop-version-floor.json";

const FETCH_TIMEOUT_MS = 10_000;

type Floor = {
  min?: string;
  latest?: string;
  downloadUrl?: string;
};

const DEFAULT_DOWNLOAD_URL =
  "https://github.com/Untrivial-ai/agent-orchestrator/releases/latest";

function usableVersion(v: string | null | undefined): string | null {
  const trimmed = v?.trim();
  return trimmed && /^\d{1,4}(\.\d{1,4}){0,3}(-[a-zA-Z0-9.]+)?$/.test(trimmed)
    ? trimmed
    : null;
}

function isBelow(running: string, floor: string): boolean {
  const r = semver.valid(semver.coerce(running));
  const f = semver.valid(semver.coerce(floor));
  if (!r || !f) return false;
  return semver.lt(r, f);
}

export async function checkDesktopVersionFloor(): Promise<void> {
  return;
}
