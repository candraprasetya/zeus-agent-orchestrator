import {
	canSubmitProjectSetup,
	ProjectSetupFormView,
	ProjectSetupHeaderView,
} from "@aoagents/product-ui";
import { useUiStore } from "../stores/ui-store";
import { useTranslation } from "react-i18next";
import * as Dialog from "@radix-ui/react-dialog";
import { useQuery } from "@tanstack/react-query";
import { ChevronLeft, FolderOpen, TriangleAlert, X, type LucideIcon } from "@/icons/hugeicons-adapter";
import { memo, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import type { components } from "../../api/schema";
import { useAgentReadinessQuery, useEnsureAgentReadiness } from "../hooks/useAgentReadinessQuery";
import { useTeams } from "../hooks/useTeams";
import { workspaceQueryOptions } from "../hooks/useWorkspaceQuery";
import { AGENT_OPTIONS, agentLabel } from "../lib/agent-options";
import {
	buildRankedAgentOptions,
	isLaunchableAgent,
	DEFAULT_AGENT_PRIORITY_RANK,
	defaultAuthorizedAgentForRole,
	type AgentInfo,
	unknownAgentReadiness,
} from "../lib/agent-select-options";
import { cn } from "../lib/utils";
import { aoBridge } from "../lib/bridge";
import { apiClient } from "../lib/api-client";
import { useAgentManagementMenu } from "../hooks/useAgentManagementMenu";
import { useSettings } from "../hooks/useSettings";
import { AgentAvatar } from "./AgentAvatar";
import { FieldDefaultHint } from "./FieldDefaultHint";
import { buildIntake, type IntakeForm, IntakeFields, intakeNeedsRule } from "./IntakeFields";
import { AgentSelectMenuItem } from "./settings/AgentSelectMenuItem";
import { SettingsRow } from "./settings/SettingsRow";
import { SettingsOptionMenu } from "./settings/SettingsOptionMenu";
import type { ProjectKind } from "../types/workspace";
import { Label } from "./ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { appI18n } from "../i18n";
import { Button } from "./ui/button";
import { Input } from "./ui/input";

type TrackerIntakeConfig = components["schemas"]["TrackerIntakeConfig"];

export type CreateProjectAgentSelection = {
	workerAgent: string;
	orchestratorAgent: string;
	trackerIntake?: TrackerIntakeConfig;
	projectType?: string;
	omniAndroidPath?: string;
	omnimonPath?: string;
	omniIosPath?: string;
	omniI18nPath?: string;
	wikiPath?: string;
	teamId?: string;
};

const EMPTY_INTAKE: IntakeForm = { enabled: false, repo: "", assignee: "" };
const AGENT_MENU_WIDTH = "w-56! min-w-56! max-w-56!";
type CreateProjectAgentSheetProps = {
	error?: string | null;
	action?: "create" | "clone";
	connected?: boolean;
	hostId?: string;
	isCreating: boolean;
	isInitializing?: boolean;
	kind: ProjectKind;
	onOpenChange: (open: boolean) => void;
	onBack?: () => void;
	onSubmit: (selection: CreateProjectAgentSelection) => Promise<void>;
	open: boolean;
	path: string | null;
	repositorySetupNeeded?: boolean;
	repositorySetupWarning?: string | null;
	shake?: boolean;
};

type SheetError = {
	title: string;
	message: string;
	tone: "warning" | "error";
};

function projectSheetError(error: string, action: "create" | "clone"): SheetError {
	const setupMessage = error.replace(/^Setup failed:\s*/i, "").trim();
	const codeMatch = setupMessage.match(/\(([A-Z0-9_]+)\)\s*$/);
	const code = codeMatch?.[1];
	const message = codeMatch ? setupMessage.slice(0, codeMatch.index).trim() : setupMessage;

	switch (code) {
		case "PROJECT_PATH_NOT_REPO_ROOT":
			return {
				title: appI18n.t("createProject.error.notRepoRootTitle"),
				message: appI18n.t("createProject.error.notRepoRootBody"),
				tone: "warning",
			};
		case "PROJECT_BARE_REPOSITORY":
			return {
				title: appI18n.t("createProject.error.bareTitle"),
				message: appI18n.t("createProject.error.bareBody"),
				tone: "warning",
			};
		case "UNSUPPORTED_GIT_REPO":
			return {
				title: appI18n.t("createProject.error.unsupportedTitle"),
				message: appI18n.t("createProject.error.unsupportedBody"),
				tone: "warning",
			};
		default:
			return {
				title: error.toLowerCase().startsWith("setup failed:")
					? appI18n.t("createProject.error.setupFailedTitle")
					: action === "clone"
						? appI18n.t("createProject.cloneFailedTitle")
						: appI18n.t("createProject.error.createFailedTitle"),
				message: message || appI18n.t("createProject.error.tryAgain"),
				tone: "error",
			};
	}
}

export function CreateProjectAgentSheet({
	action = "create",
	connected = true,
	error,
	hostId,
	isCreating,
	isInitializing = false,
	kind,
	onBack,
	onOpenChange,
	onSubmit,
	open,
	path,
	repositorySetupNeeded = false,
	repositorySetupWarning = null,
	shake = false,
}: CreateProjectAgentSheetProps) {
	const { t } = useTranslation();
	// Settings opens as a page in the center pane; step aside so it is not hidden behind this modal.
	const settingsOpen = useUiStore((state) => state.settingsModal?.scope === "global");
	const [isExiting, setIsExiting] = useState(false);
	const contentOpen = open || isExiting;
	const displayedAction = useRef(action);
	const displayedError = useRef(error);
	const displayedOnBack = useRef(onBack);
	if (open) {
		displayedAction.current = action;
		displayedError.current = error;
		displayedOnBack.current = onBack;
	}
	const agentsQuery = useAgentReadinessQuery(contentOpen && connected, hostId);
	useEnsureAgentReadiness({ enabled: contentOpen && connected, hostId });
	const agents = agentsQuery.data;
	const agentOptions = useMemo(() => agents?.agents ?? [], [agents]);
	const selectableAgents = useMemo(() => hostId ? agentOptions.filter(isLaunchableAgent) : agentOptions, [agentOptions, hostId]);
	// "configured" belongs here even though it is not a verified credential.
	// This picks the default preselection, not a gate — every agent stays
	// selectable — and excluding it would silently stop preselecting an agent
	// whose credentials AO simply cannot validate, which is most of them.
	const authorizedAgents = useMemo(
		() =>
			selectableAgents.filter((agent) =>
				["authorized", "not_applicable", "configured"].includes(agent.authentication.state),
			),
		[selectableAgents],
	);
	// Local history is an inference signal only for local projects. A remote
	// project must never infer its agent from this laptop's sessions.
	const workspacesQuery = useQuery({ ...workspaceQueryOptions, enabled: open && !hostId });
	const sessionHistory = useMemo(
		() => (workspacesQuery.data ?? []).flatMap((workspace) => workspace.sessions),
		[workspacesQuery.data],
	);
	const isLoadingAgents = connected && agents === undefined && agentsQuery.isFetching;
	const agentsError = !connected ? t("remote.connectBeforeStart") : agentsQuery.isError
		? agentsQuery.error instanceof Error
			? agentsQuery.error.message
			: t("createProject.couldNotLoadAgents")
		: null;
	const displayError = agentsError ?? (hostId && agents && selectableAgents.length === 0 ? t("agentSelector.noneReady") : null);
	const [workerAgent, setWorkerAgent] = useState("");
	const [orchestratorAgent, setOrchestratorAgent] = useState("");
	const [workerAgentTouched, setWorkerAgentTouched] = useState(false);
	const [orchestratorAgentTouched, setOrchestratorAgentTouched] = useState(false);
	const { teams } = useTeams(open && !hostId);
	const [projectType, setProjectType] = useState<string>("");
	const [teamId, setTeamId] = useState<string>("");
	const [omniAndroidPath, setOmniAndroidPath] = useState<string>("");
	const [omnimonPath, setOmnimonPath] = useState<string>("");
	const [omniIosPath, setOmniIosPath] = useState<string>("");
	const [omniI18nPath, setOmniI18nPath] = useState<string>("");
	const [wikiPath, setWikiPath] = useState<string>("");
	useEnsureAgentReadiness({
		agentIds: [workerAgent, orchestratorAgent],
		enabled: contentOpen && connected && (workerAgent !== "" || orchestratorAgent !== ""),
		hostId,
		purpose: hostId ? "launch" : "display",
	});
	const isBusy = isCreating || isInitializing;
	const [intake, setIntake] = useState<IntakeForm>(EMPTY_INTAKE);
	const { settings } = useSettings(hostId);
	const intakeVisible = !!settings?.trackerIntakeEnabled;
	const intakeIncomplete = intakeVisible && intakeNeedsRule(intake);
	const canSubmit =
		canSubmitProjectSetup({
			workerAgent,
			orchestratorAgent,
			intakeEnabled: intakeVisible && intake.enabled,
			intakeAssignee: intake.assignee,
		}) &&
		!intakeIncomplete &&
		!isBusy &&
		!isLoadingAgents && connected && (!hostId || (
			agents !== undefined && selectableAgents.some((agent) => agent.id === workerAgent) && selectableAgents.some((agent) => agent.id === orchestratorAgent)
		));
	const sheetError = displayedError.current
		? projectSheetError(displayedError.current, displayedAction.current)
		: null;
	const wasOpen = useRef(false);

	useEffect(() => {
		if (open && !wasOpen.current) {
			setWorkerAgent("");
			setOrchestratorAgent("");
			setWorkerAgentTouched(false);
			setOrchestratorAgentTouched(false);
			setIntake(EMPTY_INTAKE);
			setProjectType("");
			setTeamId("");
			setOmniAndroidPath("");
			setOmnimonPath("");
			setOmniIosPath("");
			setOmniI18nPath("");
			setWikiPath("");
		}
		wasOpen.current = open;
	}, [open]);

	useEffect(() => {
		if (open && path) {
			void checkAndAutoDetectWiki(path, wikiPath, setWikiPath);
		}
	}, [open, path]);

	useEffect(() => {
		if (!open) return;
		if (!workerAgentTouched) {
			setWorkerAgent(defaultAuthorizedAgentForRole(authorizedAgents, sessionHistory, "worker"));
		}
		if (!orchestratorAgentTouched) {
			setOrchestratorAgent(defaultAuthorizedAgentForRole(authorizedAgents, sessionHistory, "orchestrator"));
		}
	}, [authorizedAgents, open, orchestratorAgentTouched, sessionHistory, workerAgentTouched]);

	const mobileControl = (
		<div className="space-y-3 rounded-lg border border-border/60 bg-muted/20 p-3.5">
			<div className="flex items-center justify-between">
				<span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
					Mobile & Squad Configuration
				</span>
			</div>

			<div className="grid gap-3 sm:grid-cols-2">
				<div className="space-y-1.5">
					<Label className="text-xs font-medium text-foreground">Project Type</Label>
					<SettingsOptionMenu
						aria-label="Project Type"
						value={projectType}
						options={[
							{ value: "", label: "None (Not specified)" },
							{ value: "android", label: "Android" },
							{ value: "ios", label: "iOS" },
						]}
						placeholder="Select Project Type"
						triggerClassName="w-full justify-between"
						onChange={(type) => {
							setProjectType(type);
							if (type === "ios" && !teamId) setTeamId("zeus-team-ios");
							else if (type === "android" && !teamId) setTeamId("zeus-team-android");
						}}
					/>
				</div>

				<div className="space-y-1.5">
					<Label className="text-xs font-medium text-foreground">Assigned Squad</Label>
					<SettingsOptionMenu
						aria-label="Assigned Squad"
						value={teamId}
						options={[
							{ value: "", label: "None (No squad)" },
							...teams.map((t) => ({ value: t.id, label: t.name })),
						]}
						placeholder="Select Squad"
						triggerClassName="w-full justify-between"
						onChange={(selectedTeam) => {
							setTeamId(selectedTeam);
							if (selectedTeam === "zeus-team-ios" && !projectType) setProjectType("ios");
							else if (selectedTeam === "zeus-team-android" && !projectType) setProjectType("android");
						}}
					/>
				</div>
			</div>

			{projectType === "android" && (
				<div className="space-y-2.5 pt-1">
					<div className="space-y-1.5">
						<Label className="text-xs font-medium text-foreground">Omni-Android folder</Label>
						<FolderBrowseInput
							value={omniAndroidPath}
							placeholder="/path/to/Omni-Android"
							onChange={(val) => {
								setOmniAndroidPath(val);
								void checkAndAutoDetectWiki(val, wikiPath, setWikiPath);
							}}
							onBrowse={async () => {
								const picked = await aoBridge.app.chooseDirectory({
									title: "Select Omni-Android Directory",
									defaultPath: omniAndroidPath || path || undefined,
								});
								if (picked) {
									setOmniAndroidPath(picked);
									void checkAndAutoDetectWiki(picked, wikiPath, setWikiPath);
								}
							}}
						/>
					</div>
					<div className="space-y-1.5">
						<Label className="text-xs font-medium text-foreground">Omnimon folder</Label>
						<FolderBrowseInput
							value={omnimonPath}
							placeholder="/path/to/Omnimon"
							onChange={setOmnimonPath}
							onBrowse={async () => {
								const picked = await aoBridge.app.chooseDirectory({
									title: "Select Omnimon Directory",
									defaultPath: omnimonPath || undefined,
								});
								if (picked) setOmnimonPath(picked);
							}}
						/>
					</div>
				</div>
			)}

			{projectType === "ios" && (
				<div className="space-y-2.5 pt-1">
					<div className="space-y-1.5">
						<Label className="text-xs font-medium text-foreground">Omni-IOS folder</Label>
						<FolderBrowseInput
							value={omniIosPath}
							placeholder="/path/to/Omni-IOS"
							onChange={(val) => {
								setOmniIosPath(val);
								void checkAndAutoDetectWiki(val, wikiPath, setWikiPath);
							}}
							onBrowse={async () => {
								const picked = await aoBridge.app.chooseDirectory({
									title: "Select Omni-IOS Directory",
									defaultPath: omniIosPath || path || undefined,
								});
								if (picked) {
									setOmniIosPath(picked);
									void checkAndAutoDetectWiki(picked, wikiPath, setWikiPath);
								}
							}}
						/>
					</div>
				</div>
			)}

			{(projectType === "ios" || projectType === "android") && (
				<div className="space-y-2.5 pt-1">
					<div className="space-y-1.5">
						<Label className="text-xs font-medium text-foreground">omni-i18n folder</Label>
						<FolderBrowseInput
							value={omniI18nPath}
							placeholder="/path/to/omni-i18n"
							onChange={setOmniI18nPath}
							onBrowse={async () => {
								const picked = await aoBridge.app.chooseDirectory({
									title: "Select omni-i18n Directory",
									defaultPath: omniI18nPath || undefined,
								});
								if (picked) setOmniI18nPath(picked);
							}}
						/>
					</div>
					<div className="space-y-1.5">
						<Label className="text-xs font-medium text-foreground">Wiki folder</Label>
						<FolderBrowseInput
							value={wikiPath}
							placeholder="Auto-detected or select wiki folder..."
							onChange={setWikiPath}
							onBrowse={async () => {
								const picked = await aoBridge.app.chooseDirectory({
									title: "Select Wiki Directory",
									defaultPath: wikiPath || undefined,
								});
								if (picked) setWikiPath(picked);
							}}
						/>
					</div>
				</div>
			)}
		</div>
	);

	if (isBusy) return null;

	return (
		<Dialog.Root
			open={open && !settingsOpen}
			onOpenChange={(next) => {
				if (isBusy) return;
				setIsExiting(!next);
				onOpenChange(next);
			}}
		>
			<Dialog.Portal>
				<Dialog.Content
					className={cn("fixed left-1/2 top-1/2 z-overlay w-dialog-lg -translate-x-1/2 -translate-y-1/2 overflow-hidden rounded-lg border border-border bg-popover p-0 text-popover-foreground shadow-xl data-[state=open]:animate-modal-in data-[state=closed]:animate-modal-out motion-reduce:animate-none", shake && "modal-shake")}
					onAnimationEnd={(event) => {
						if (!open && event.target === event.currentTarget) setIsExiting(false);
					}}
				>
					<ProjectSetupHeaderView
						CloseButton={ProjectSheetCloseButton}
						Description={Dialog.Description}
						Title={Dialog.Title}
						closeIcon={<X className="size-icon-base" aria-hidden="true" />}
						closeLabel={t("createProject.closeAgents")}
						disabled={isBusy}
						leadingAction={
							displayedOnBack.current ? (
								<Button
									type="button"
									variant="outline"
									size="icon"
									aria-label={t("createProject.cloneBackToDetails")}
									disabled={isBusy}
								onClick={displayedOnBack.current}
								>
									<ChevronLeft className="size-4" aria-hidden="true" />
								</Button>
							) : undefined
						}
						path={path ?? ""}
						showPath={false}
						title={
							kind === "workspace"
								? t("createProject.setupWorkspace")
								: t("createProject.setupProject")
						}
					/>
					<ProjectSetupFormView
						agentControls={{
							worker: (
								<RequiredAgentField
									id="newProjectWorkerAgent"
									label={t("createProject.workerAgent")}
									placeholder={t("createProject.selectWorker")}
									value={workerAgent}
									agents={selectableAgents}
									disabled={isLoadingAgents}
									labelClassName="agents-sheet-label"
									triggerClassName="agents-sheet-control"
									contentClassName="agents-sheet-menu"
									hostId={hostId}
									onChange={(value) => {
										setWorkerAgent(value);
										setWorkerAgentTouched(true);
									}}
								/>
							),
							orchestrator: (
								<RequiredAgentField
									id="newProjectOrchestratorAgent"
									label={t("createProject.orchestratorAgent")}
									placeholder={t("createProject.selectOrchestrator")}
									value={orchestratorAgent}
									agents={selectableAgents}
									disabled={isLoadingAgents}
									labelClassName="agents-sheet-label"
									triggerClassName="agents-sheet-control"
									contentClassName="agents-sheet-menu"
									hostId={hostId}
									onChange={(value) => {
										setOrchestratorAgent(value);
										setOrchestratorAgentTouched(true);
									}}
								/>
							),
						}}
						agents={{
							error: displayError,
							loading: isLoadingAgents,
							loadingMessage: t("createProject.loadingAgents"),
							onRetry: () => void agentsQuery.refetch(),
							retrying: agentsQuery.isFetching,
							retryLabel: t("createProject.retry"),
						}}
						alert={
							sheetError
								? {
										...sheetError,
										icon: (
											<TriangleAlert
												className={
													sheetError.tone === "warning"
														? "mt-0.5 size-icon-sm shrink-0 text-warning"
														: "mt-0.5 size-icon-sm shrink-0 text-destructive"
												}
												aria-hidden="true"
											/>
										),
									}
								: null
						}
						canSubmit={canSubmit}
						intakeControl={
							intakeVisible ? (
								<IntakeFields
									form={intake}
									onChange={(patch) => setIntake((f) => ({ ...f, ...patch }))}
									compact
									controlClassName="agents-sheet-control"
									labelClassName="agents-sheet-label"
								/>
							) : null
						}
						extraControl={mobileControl}
						isBusy={isBusy}
						onCancel={() => onOpenChange(false)}
						onSubmit={() =>
							void onSubmit({
								workerAgent,
								orchestratorAgent,
								trackerIntake: intakeVisible ? buildIntake(intake) : undefined,
								projectType: projectType || undefined,
								omniAndroidPath: omniAndroidPath || undefined,
								omnimonPath: omnimonPath || undefined,
								omniIosPath: omniIosPath || undefined,
								omniI18nPath: omniI18nPath || undefined,
								wikiPath: wikiPath || undefined,
								teamId: teamId || undefined,
							})
						}
						setupNotice={
							repositorySetupNeeded
								? { message: t("createProject.gitSetupNotice"), warning: repositorySetupWarning }
								: null
						}
						submitLabel={
							isInitializing
								? t("createProject.settingUp")
								: isCreating
									? action === "clone"
										? t("createProject.cloning")
										: t("createProject.creating")
									: action === "clone"
										? t("createProject.clone")
										: kind === "workspace"
											? t("createProject.createWorkspaceAndStart")
											: t("createProject.createAndStart")
						}
						submitClassName={cn(
							"inline-flex h-control-form items-center gap-2 rounded-md bg-primary px-3 text-sm text-primary-foreground hover:bg-primary/80",
							(isCreating || isInitializing) && "before:size-3.5 before:shrink-0 before:animate-spin before:rounded-full before:border-2 before:border-current before:border-r-transparent before:content-['']",
						)}
					/>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	);
}

function ProjectSheetCloseButton({
	children,
	disabled,
	"aria-label": ariaLabel,
}: {
	children: ReactNode;
	disabled: boolean;
	"aria-label": string;
}) {
	return (
		<Dialog.Close asChild>
			<button
				type="button"
				className="settings-close-button"
				aria-label={ariaLabel}
				disabled={disabled}
			>
				{children}
			</button>
		</Dialog.Close>
	);
}

export const RequiredAgentField = memo(function RequiredAgentField({
	agents,
	disabled = false,
	hint,
	icon,
	id,
	invalid = false,
	label,
	onChange,
	placeholder,
	hostId,
	manageAgents = true,
	manageView = "local",
	triggerClassName,
	managementLabel: managementLabelOverride,
	labelClassName,
	contentClassName,
	value,
	variant = "stacked",
}: {
	agents?: AgentInfo[];
	disabled?: boolean;
	/** Caption beside the label, e.g. naming where a preselected default came from. */
	hint?: string;
	icon?: LucideIcon;
	id: string;
	invalid?: boolean;
	label: string;
	onChange: (value: string) => void;
	placeholder: string;
	hostId?: string;
	/** Cloud tasks use remote availability, not this computer's Harness settings. */
	manageAgents?: boolean;
	/** Which Harness settings view "manage" opens: local logins or cloud connections. */
	manageView?: "local" | "cloud";
	triggerClassName?: string;
	/** Optional shorter action copy for compact selectors. */
	managementLabel?: string;
	labelClassName?: string;
	contentClassName?: string;
	value: string;
	variant?: "stacked" | "settings-row" | "settings-control" | "chip";
}) {
	const { t } = useTranslation();
	const fallbackAgents: AgentInfo[] = AGENT_OPTIONS.map((agent) => unknownAgentReadiness(agent, agentLabel(agent)));
	const options = buildRankedAgentOptions({
		agents,
		priorityRank: DEFAULT_AGENT_PRIORITY_RANK,
		fallbackAgents,
	});

	const selectedOption = options.find((agent) => agent.id === value) ?? (value ? unknownAgentReadiness(value, agentLabel(value)) : undefined);
	const hasReadinessSnapshot = agents !== undefined;
	const needsSetup = manageAgents && hasReadinessSnapshot && Boolean(selectedOption && !isLaunchableAgent(selectedOption));
	const visibleOptions = manageAgents && hasReadinessSnapshot ? options.filter(isLaunchableAgent) : options;
	// Local is Harness settings' default view, so only cloud needs to ask for one.
	const management = useAgentManagementMenu(needsSetup ? value : undefined, hostId, manageView === "cloud" ? "cloud" : undefined);
	const manageLabel = managementLabelOverride ?? (manageView === "cloud" ? t("agentSelector.manageCloud") : t("agentSelector.manage"));
	const managementAction = manageAgents ? { label: manageLabel, onSelect: management.requestManagement } : undefined;
	const setupHint = needsSetup ? <span className="text-xs text-muted-foreground">{t("agentSelector.needsSetup")}</span> : null;

	if (variant === "settings-row" || variant === "settings-control") {
		const menuOptions = visibleOptions.map((agent) => ({
			value: agent.id,
			label: agent.label,
			disabled: agent.disabled,
		}));

		const control = (
				<SettingsOptionMenu
					aria-label={label}
					value={value}
					placeholder={placeholder}
					options={menuOptions}
					action={managementAction}
					emptyLabel={manageAgents ? t("agentSelector.noneReady") : undefined}
					triggerRef={management.triggerRef}
					onCloseAutoFocus={management.onCloseAutoFocus}
					disabled={disabled}
					onChange={onChange}
					triggerClassName={cn(variant === "settings-control" && "w-fit", invalid && "text-error")}
					menuClassName={cn("settings-agent-menu-surface", AGENT_MENU_WIDTH)}
					menuItemClassName="settings-agent-menu-item"
					renderTrigger={() => (
						<span className="flex min-w-0 items-center gap-2">
							{selectedOption ? <AgentAvatar provider={selectedOption.id} className="size-icon-lg" /> : null}
							<span className="min-w-0 truncate">{selectedOption?.label ?? placeholder}</span>
							{setupHint}
						</span>
					)}
					renderMenuItem={(option, selected) => {
						const agent = options.find((entry) => entry.id === option.value);
						if (!agent) return option.label;
						return (
							<AgentSelectMenuItem
								agentId={agent.id}
								label={agent.label}
								selected={selected}
								status={agent.status}
								statusTone={agent.statusTone}
								disabled={agent.disabled}
							/>
						);
					}}
				/>
		);
		return variant === "settings-row" ? <SettingsRow icon={icon} label={label}>{control}</SettingsRow> : control;
	}

	// Chip: the value reads as part of a sentence ("Runs with Codex") rather than
	// as a form field, so the label is carried by that sentence, not by a <Label>.
	// Built on the same SettingsOptionMenu as the settings-row variant (and the
	// model chip beside it) so both halves of the pill share one dropdown
	// component instead of a Select-based menu and a DropdownMenu-based one.
	if (variant === "chip") {
		const menuOptions = visibleOptions.map((agent) => ({
			value: agent.id,
			label: agent.label,
			disabled: agent.disabled,
		}));

		return (
			<SettingsOptionMenu
				aria-label={label}
				value={value}
				placeholder={placeholder}
				options={menuOptions}
				action={managementAction}
				emptyLabel={manageAgents ? t("agentSelector.noneReady") : undefined}
				triggerRef={management.triggerRef}
				onCloseAutoFocus={management.onCloseAutoFocus}
				disabled={disabled}
				onChange={onChange}
				menuAlign="start"
				triggerClassName={cn(
					"composer-chip composer-toolbar-option w-full justify-between",
					invalid && "text-error",
					triggerClassName,
				)}
				menuClassName={cn(
					AGENT_MENU_WIDTH,
					contentClassName,
				)}
				renderTrigger={() => (
					<span className="flex min-w-0 items-center gap-2">
						{selectedOption ? (
							<AgentAvatar provider={selectedOption.id} className="size-icon-base" decorative />
						) : null}
						<span className="min-w-0 truncate text-control text-foreground" title={selectedOption?.label ?? placeholder}>
							{selectedOption?.label ?? placeholder}
						</span>
						{setupHint}
					</span>
				)}
				renderMenuItem={(option, selected) => {
					const agent = options.find((entry) => entry.id === option.value);
					if (!agent) return option.label;
					return (
						<AgentSelectMenuItem
							agentId={agent.id}
							label={agent.label}
							selected={selected}
							status={agent.status}
							statusTone={agent.statusTone}
							disabled={agent.disabled}
						/>
					);
				}}
			/>
		);
	}

	return (
		<div className="flex flex-col gap-1.5">
			<div className="flex min-w-0 items-baseline gap-1.5">
				<Label htmlFor={id} className={cn("text-xs font-medium text-muted-foreground", labelClassName)}>
					{label}
				</Label>
				{hint && <FieldDefaultHint text={hint} />}
			</div>
			<Select value={value} onValueChange={(next) => {
				if (manageAgents && next === "__manage_agents__") management.requestManagement();
				else onChange(next);
			}} disabled={disabled}>
				<SelectTrigger
					ref={management.triggerRef}
					id={id}
					size="sm"
					className={cn("w-full text-control", triggerClassName)}
					aria-label={label}
					aria-invalid={invalid || undefined}
				>
					{/* Radix would otherwise clone the whole menu row into the trigger,
					    dragging the selected checkmark and install status with it. */}
					<SelectValue placeholder={placeholder}>
						{selectedOption ? (
							<span className="flex min-w-0 items-center gap-3">
								<AgentAvatar provider={selectedOption.id} className="size-icon-lg" decorative />
								<span className="min-w-0 truncate">{selectedOption.label}</span>
								{setupHint}
							</span>
						) : null}
					</SelectValue>
				</SelectTrigger>
				<SelectContent
					onCloseAutoFocus={management.onCloseAutoFocus}
					position="popper"
					side="bottom"
					align="start"
					sideOffset={4}
					className={cn("max-h-select-menu-max!", contentClassName)}
				>
					{visibleOptions.map((agent) => (
						<SelectItem
							key={agent.id}
							value={agent.id}
							disabled={agent.disabled}
							className="[&>span:last-child]:w-full"
						>
							<AgentSelectMenuItem
								agentId={agent.id}
								label={agent.label}
								selected={value === agent.id}
								status={agent.status}
								statusTone={agent.statusTone}
								disabled={agent.disabled}
							/>
						</SelectItem>
					))}
					{manageAgents && visibleOptions.length === 0 && <p className="px-2 py-1.5 text-xs text-muted-foreground">{t("agentSelector.noneReady")}</p>}
					{manageAgents && <SelectItem value="__manage_agents__" className="mt-1 border-t border-border">{manageLabel}</SelectItem>}
				</SelectContent>
			</Select>
		</div>
	);
});

function FolderBrowseInput({
	value,
	placeholder,
	onChange,
	onBrowse,
}: {
	value: string;
	placeholder?: string;
	onChange: (value: string) => void;
	onBrowse: () => void;
}) {
	return (
		<div className="flex w-full items-center gap-1.5">
			<Input
				value={value}
				placeholder={placeholder}
				onChange={(e) => onChange(e.target.value)}
				className="h-8 text-xs font-mono bg-background/50 flex-1"
			/>
			<Button
				type="button"
				variant="outline"
				size="sm"
				className="h-8 shrink-0 px-2.5 text-xs gap-1.5"
				onClick={onBrowse}
			>
				<FolderOpen className="size-3.5" />
				Browse
			</Button>
		</div>
	);
}

async function checkAndAutoDetectWiki(
	rootPath: string,
	currentWiki: string,
	onFound: (wikiPath: string) => void,
) {
	if (!rootPath.trim() || currentWiki.trim()) return;
	try {
		const { data } = await apiClient.GET("/api/v1/fs/dirs", {
			params: { query: { path: rootPath.trim() } },
		});
		if (data?.entries) {
			const found = data.entries.find((e) => e.name.toLowerCase() === "wiki");
			if (found) {
				onFound(found.path);
			}
		}
	} catch {
		// Ignore listing errors
	}
}
