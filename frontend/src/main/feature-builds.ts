import { app } from "electron";
import type { UpdateSettings } from "./update-settings";

export interface FeatureBuild {
	pr: number;
	title: string;
	base: string;
	sha: string;
	slug: string;
	/** The version/tag of the build (e.g. "1.2.3-pr2270.0"). */
	buildId: string;
	publishedAt: string;
}

/**
 * Parse a version string for a feature-build prerelease identifier.
 * Matches "-pr<N>.<12-digit-ts>" (with optional leading "v"). Returns { pr } or
 * null. The anchor is kept in sync with frontend/scripts/feature-version.mjs's
 * parser (a local copy avoids a cross-dir import from the main process into the
 * build-scripts directory).
 */
export function parseFeatureBuild(version: string): { pr: number } | null {
	const m = version.match(/-pr(\d+)\.\d{12}/);
	if (!m) return null;
	const pr = parseInt(m[1], 10);
	return Number.isFinite(pr) && pr > 0 ? { pr } : null;
}

/** Return the feature-build pin for the currently running app version, or null. */
export function getActiveFeatureBuild(): { pr: number } | null {
	return parseFeatureBuild(app.getVersion());
}

async function collectFeatureBuilds(): Promise<FeatureBuild[]> {
	return [];
}

/**
 * List available feature builds. Never throws: returns [] on network/HTTP
 * errors (the renderer just shows an empty picker). See collectFeatureBuilds
 * for the filter/group semantics.
 */
export async function listFeatureBuilds(): Promise<FeatureBuild[]> {
	try {
		return await collectFeatureBuilds();
	} catch (err) {
		console.warn("[feature-builds] failed to list feature builds:", err);
		return [];
	}
}

/**
 * Reconcile a pinned feature build against the live set. If the pinned PR no
 * longer has a live build (merged, closed, deleted, or expired past 7 days),
 * return settings with the pin cleared so the caller can fall back to the home
 * channel. A fetch failure is treated as "keep the pin" so a transient network
 * error never strands the user off their pinned build.
 */
export async function reconcileFeaturePin(
	settings: UpdateSettings,
): Promise<{ settings: UpdateSettings; cleared: boolean }> {
	if (!settings.feature) return { settings, cleared: false };
	let builds: FeatureBuild[];
	try {
		builds = await collectFeatureBuilds();
	} catch {
		return { settings, cleared: false };
	}
	if (builds.some((b) => b.pr === settings.feature!.pr)) return { settings, cleared: false };
	return { settings: { ...settings, feature: null }, cleared: true };
}
