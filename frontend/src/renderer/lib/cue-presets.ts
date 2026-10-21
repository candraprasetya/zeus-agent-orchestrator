import type { CueInput } from "./cues";

export type CuePreset = {
	id: string;
	platform: "android" | "ios";
	name: string;
	type: "command" | "agent";
	command?: string;
	prompt?: string;
	description: string;
};

export const ANDROID_CUE_PRESETS: CuePreset[] = [
	{
		id: "android-build-debug",
		platform: "android",
		name: "Android Build (Debug)",
		type: "command",
		command: "./gradlew assembleDebug",
		description: "Compile and assemble the Android debug APK",
	},
	{
		id: "android-test",
		platform: "android",
		name: "Android Test",
		type: "command",
		command: "./gradlew test",
		description: "Run Android local unit tests",
	},
	{
		id: "android-lint",
		platform: "android",
		name: "Android Lint",
		type: "command",
		command: "./gradlew lint",
		description: "Run Android Lint checks for potential bugs and quality issues",
	},
	{
		id: "android-clean",
		platform: "android",
		name: "Android Clean",
		type: "command",
		command: "./gradlew clean",
		description: "Clean Gradle build cache and generated directories",
	},
	{
		id: "android-agent-fix",
		platform: "android",
		name: "Fix Android Issues",
		type: "agent",
		prompt: "Diagnose and fix any Android compilation errors, Gradle configuration issues, or Android Lint warnings in this project.",
		description: "Instruct agent to fix Android build and lint issues",
	},
	{
		id: "android-agent-review",
		platform: "android",
		name: "Android Code Review",
		type: "agent",
		prompt: "Review recent Android changes for Clean Architecture / MVVM / MVI patterns, Jetpack Compose / Kotlin conventions, memory leaks, and performance.",
		description: "Instruct agent to perform Android architecture and code review",
	},
];

export const IOS_CUE_PRESETS: CuePreset[] = [
	{
		id: "ios-build",
		platform: "ios",
		name: "iOS Build",
		type: "command",
		command: "xcodebuild -destination 'generic/platform=iOS Simulator' build",
		description: "Build iOS target for simulator",
	},
	{
		id: "ios-test",
		platform: "ios",
		name: "iOS Test",
		type: "command",
		command: "xcodebuild test",
		description: "Run iOS unit and UI tests",
	},
	{
		id: "ios-swiftlint",
		platform: "ios",
		name: "SwiftLint",
		type: "command",
		command: "swiftlint",
		description: "Run SwiftLint code style and convention checks",
	},
	{
		id: "ios-pod-install",
		platform: "ios",
		name: "CocoaPods Install",
		type: "command",
		command: "pod install",
		description: "Install CocoaPods dependencies",
	},
	{
		id: "ios-agent-fix",
		platform: "ios",
		name: "Fix iOS Issues",
		type: "agent",
		prompt: "Diagnose and fix any Swift compilation errors, Xcode project misconfigurations, or SwiftLint violations in this project.",
		description: "Instruct agent to fix iOS build, Swift, and Xcode issues",
	},
	{
		id: "ios-agent-review",
		platform: "ios",
		name: "iOS Code Review",
		type: "agent",
		prompt: "Review recent iOS changes for Swift concurrency, SwiftUI/UIKit best practices, retain cycles, and architecture patterns.",
		description: "Instruct agent to perform iOS architecture and code review",
	},
];

export const ALL_CUE_PRESETS: CuePreset[] = [...ANDROID_CUE_PRESETS, ...IOS_CUE_PRESETS];

export function getPresetsForPlatform(platform: "android" | "ios"): CuePreset[] {
	return platform === "android" ? ANDROID_CUE_PRESETS : IOS_CUE_PRESETS;
}

export function presetToCueInput(preset: CuePreset): CueInput {
	return {
		name: preset.name,
		type: preset.type,
		...(preset.type === "command" ? { command: preset.command } : { prompt: preset.prompt }),
	};
}
