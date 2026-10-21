import { useTranslation } from "react-i18next";
import { useEffect, useId, useRef, useState, type ReactNode } from "react";
import {
	collectReportProblemDiagnostics,
	formatReportProblemDraft,
	reportProblemDestinationUrl,
	type ReportProblemDiagnostics,
	type ReportProblemEmailProvider,
	type ReportProblemOutput,
} from "../../lib/report-problem";
import { aoBridge } from "../../lib/bridge";
import { isWindowsPlatform } from "../../lib/platform";
import { captureRendererEvent } from "../../lib/telemetry";
import { Button } from "../ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "../ui/dropdown-menu";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";

import { HugeiconsIcon } from "@hugeicons/react";
import { DiscordIcon as HugeDiscord, GithubIcon as HugeGithub, Mail01Icon as HugeMail } from "@hugeicons/core-free-icons";

type DestinationIconProps = {
	className?: string;
};

function GithubIcon({ className }: DestinationIconProps) {
	return <HugeiconsIcon icon={HugeGithub} className={className} size={16} aria-hidden="true" />;
}

function DiscordIcon({ className }: DestinationIconProps) {
	return <HugeiconsIcon icon={HugeDiscord} className={className} size={16} aria-hidden="true" />;
}

function EmailIcon({ className }: DestinationIconProps) {
	return <HugeiconsIcon icon={HugeMail} className={className} size={16} aria-hidden="true" />;
}

const DEFAULT_DIAGNOSTICS: ReportProblemDiagnostics = {
	appVersion: "unknown",
	buildMode: "unknown",
	daemonState: "unknown",
	generatedAt: "unknown",
	platform: "unknown",
	routeSurface: "unknown",
};

type DestinationOption = {
	value: ReportProblemOutput;
	label: string;
	action: string;
	icon: (props: DestinationIconProps) => ReactNode;
};

export function ReportProblemContent({ active }: { active: boolean }) {
	const { t } = useTranslation();
	const destinations: DestinationOption[] = [
		{ value: "github", label: t("report.github"), action: t("report.githubAction"), icon: GithubIcon },
		{ value: "discord", label: t("report.discord"), action: t("report.discordAction"), icon: DiscordIcon },
		{ value: "email", label: t("report.email"), action: t("report.emailAction"), icon: EmailIcon },
	];
	const titleId = useId();
	const detailsId = useId();
	const titleRef = useRef<HTMLInputElement>(null);
	const [summary, setSummary] = useState("");
	const [details, setDetails] = useState("");
	const [copiedOutput, setCopiedOutput] = useState<ReportProblemOutput | null>(null);
	const [copyError, setCopyError] = useState<string | null>(null);
	const [diagnostics, setDiagnostics] = useState<ReportProblemDiagnostics>(DEFAULT_DIAGNOSTICS);

	const copiedLabel = destinations.find((option) => option.value === copiedOutput)?.label;

	useEffect(() => {
		if (!active) {
			setSummary("");
			setDetails("");
			setCopiedOutput(null);
			setCopyError(null);
			return;
		}
		void captureRendererEvent("ao.renderer.support_opened");
		let cancelled = false;
		void collectReportProblemDiagnostics().then((nextDiagnostics) => {
			if (!cancelled) setDiagnostics(nextDiagnostics);
		});
		return () => {
			cancelled = true;
		};
	}, [active]);

	const input = { summary, details };
	const canSubmit = summary.trim().length > 0 && details.trim().length > 0;

	const clearStatus = () => {
		setCopiedOutput(null);
		setCopyError(null);
	};

	const copyDraft = async (output: ReportProblemOutput, emailProvider: ReportProblemEmailProvider = "system") => {
		if (!canSubmit) return;
		setCopyError(null);
		const draft = formatReportProblemDraft(input, diagnostics, output);
		try {
			await aoBridge.clipboard.writeText(draft);
			const destinationUrl = reportProblemDestinationUrl(input, diagnostics, output, emailProvider);
			if (destinationUrl) {
				await aoBridge.app.openExternal(destinationUrl);
			}
			setCopiedOutput(output);
			setSummary("");
			setDetails("");
			void captureRendererEvent("ao.renderer.support_submitted", { destination: output, outcome: "succeeded" });
		} catch (err) {
			setCopyError(err instanceof Error ? err.message : t("report.copyFailed"));
			setCopiedOutput(null);
			void captureRendererEvent("ao.renderer.support_submitted", { destination: output, outcome: "failed" });
		}
	};

	return (
		<div
			className="flex flex-col gap-4"
			onKeyDown={(event) => {
				if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
					event.preventDefault();
					void copyDraft("github");
				}
			}}
		>
			<div className="flex flex-col gap-1.5">
				<label className="settings-field-label" htmlFor={titleId}>
					{t("report.titleLabel")}
				</label>
				<input
					ref={titleRef}
					id={titleId}
					className="settings-field-control h-control-form rounded-md!"
					value={summary}
					onChange={(event) => {
						setSummary(event.target.value);
						clearStatus();
					}}
					placeholder={t("report.titlePlaceholder")}
				/>
			</div>

			<div className="flex flex-col gap-1.5">
				<label className="settings-field-label" htmlFor={detailsId}>
					{t("report.whatHappened")}
				</label>
				<textarea
					id={detailsId}
					className="settings-field-control min-h-(--size-textarea-min) resize-none overflow-y-auto py-2.5 rounded-md!"
					value={details}
					onChange={(event) => {
						setDetails(event.target.value);
						clearStatus();
					}}
					placeholder={t("report.detailsPlaceholder")}
				/>
			</div>

			<div role="group" aria-label={t("report.destination")} className="flex flex-wrap items-center justify-end gap-3">
				{copyError ? (
					<p role="alert" className="text-caption leading-4 text-error">
						{copyError}
					</p>
				) : null}
				{copiedLabel && !copyError ? (
					<p className="text-caption leading-4 text-success">{t("report.draftCopied", { label: copiedLabel })}</p>
				) : null}
				<div className="flex items-center gap-1.5">
					{destinations.map((option) => {
						if (option.value === "email" && isWindowsPlatform()) {
							return (
								<DropdownMenu key={option.value}>
									<Tooltip>
										<TooltipTrigger asChild>
											<DropdownMenuTrigger asChild>
												<Button
													type="button"
													variant="ghost"
													size="icon-sm"
													className="text-settings-muted hover:text-settings-label"
													disabled={!canSubmit}
													aria-label={option.action}
												>
													<option.icon className="size-icon-sm" aria-hidden="true" />
												</Button>
											</DropdownMenuTrigger>
										</TooltipTrigger>
										<TooltipContent>{option.action}</TooltipContent>
									</Tooltip>
									<DropdownMenuContent align="end" side="top" className="w-52">
										<DropdownMenuItem onSelect={() => void copyDraft("email", "system")}>
											{t("report.emailSystem")}
										</DropdownMenuItem>
										<DropdownMenuItem onSelect={() => void copyDraft("email", "gmail")}>
											{t("report.emailGmail")}
										</DropdownMenuItem>
										<DropdownMenuItem onSelect={() => void copyDraft("email", "outlook")}>
											{t("report.emailOutlook")}
										</DropdownMenuItem>
									</DropdownMenuContent>
								</DropdownMenu>
							);
						}
						return (
							<Tooltip key={option.value}>
								<TooltipTrigger asChild>
									<Button
										type="button"
										variant="ghost"
										size="icon-sm"
										className="text-settings-muted hover:text-settings-label"
										disabled={!canSubmit}
										aria-label={option.action}
										onClick={() => {
											if (!canSubmit) return;
											void copyDraft(option.value);
										}}
									>
										<option.icon className="size-icon-sm" aria-hidden="true" />
									</Button>
								</TooltipTrigger>
								<TooltipContent>{option.action}</TooltipContent>
							</Tooltip>
						);
					})}
				</div>
			</div>
		</div>
	);
}
