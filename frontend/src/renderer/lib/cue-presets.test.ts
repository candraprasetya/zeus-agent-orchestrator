import { describe, expect, it } from "vitest";
import {
	ALL_CUE_PRESETS,
	ANDROID_CUE_PRESETS,
	IOS_CUE_PRESETS,
	getPresetsForPlatform,
	presetToCueInput,
} from "./cue-presets";
import { CUE_LIMITS } from "./cues";

describe("cue-presets", () => {
	it("contains non-empty presets for android and ios", () => {
		expect(ANDROID_CUE_PRESETS.length).toBeGreaterThan(0);
		expect(IOS_CUE_PRESETS.length).toBeGreaterThan(0);
		expect(ALL_CUE_PRESETS.length).toBe(ANDROID_CUE_PRESETS.length + IOS_CUE_PRESETS.length);
	});

	it("resolves presets by platform correctly", () => {
		expect(getPresetsForPlatform("android")).toEqual(ANDROID_CUE_PRESETS);
		expect(getPresetsForPlatform("ios")).toEqual(IOS_CUE_PRESETS);
	});

	it("converts command and agent presets to valid CueInput adhering to limits", () => {
		const encoder = new TextEncoder();
		for (const preset of ALL_CUE_PRESETS) {
			const input = presetToCueInput(preset);
			expect(input.name).toBe(preset.name);
			expect(input.type).toBe(preset.type);
			expect(encoder.encode(input.name).length).toBeLessThanOrEqual(CUE_LIMITS.name);

			if (preset.type === "command") {
				expect(input.command).toBe(preset.command);
				expect(input.command).toBeTruthy();
				expect(encoder.encode(input.command!).length).toBeLessThanOrEqual(CUE_LIMITS.command);
			} else {
				expect(input.prompt).toBe(preset.prompt);
				expect(input.prompt).toBeTruthy();
				expect(encoder.encode(input.prompt!).length).toBeLessThanOrEqual(CUE_LIMITS.prompt);
			}
		}
	});
});
