import { FolderOpen, Sparkles, Terminal } from "@/icons/hugeicons-adapter";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useImportSkills, useSkills } from "../../hooks/useSkills";
import { aoBridge } from "../../lib/bridge";
import { SettingsRow } from "./SettingsRow";
import { SettingsSection } from "./SettingsSection";

export function SkillsSection({ titleHidden }: { titleHidden?: boolean }) {
	const { t } = useTranslation();
	const { skills, isLoading } = useSkills();
	const { importSkills, isPending: isImporting } = useImportSkills();
	const [statusMsg, setStatusMsg] = useState<string | null>(null);

	const handleChooseAndImport = async () => {
		try {
			const selectedPath = await aoBridge.app.chooseDirectory("Select Skills Directory");
			if (!selectedPath) return;

			setStatusMsg(t("settings.skills.importing", "Importing skills..."));
			const imported = await importSkills(selectedPath);
			setStatusMsg(
				t("settings.skills.importedSuccess", "Successfully imported {{count}} skill(s)", {
					count: imported.length,
				}),
			);
			setTimeout(() => setStatusMsg(null), 4000);
		} catch (err) {
			setStatusMsg(err instanceof Error ? err.message : String(err));
		}
	};

	return (
		<>
			<SettingsSection
				title={t("settings.skills.title", "Skills & Packages")}
				sectionId="skills"
				titleHidden={titleHidden}
				grouped
				action={
					<button
						type="button"
						onClick={() => void handleChooseAndImport()}
						disabled={isImporting}
						className="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-medium bg-primary text-primary-foreground rounded-full hover:opacity-90 disabled:opacity-50 transition shadow-sm"
					>
						<FolderOpen className="w-3.5 h-3.5" />
						{isImporting
							? t("settings.skills.importingBtn", "Importing...")
							: t("settings.skills.importBtn", "Import Skills...")}
					</button>
				}
			>
				<SettingsRow
					label={t("settings.skills.manageLabel", "Installed Skills")}
					description={
						statusMsg ??
						t(
							"settings.skills.manageDesc",
							"Skills available to AI agents (Mage, OpenCode, Codex) in this environment.",
						)
					}
				>
					<span className="text-xs text-muted-foreground font-mono">
						{skills.length} {t("settings.skills.count", "skills")}
					</span>
				</SettingsRow>
			</SettingsSection>

			<SettingsSection
				title={t("settings.skills.catalogTitle", "Skill Catalog")}
				grouped
			>
				{isLoading ? (
					<div className="p-4 text-xs text-muted-foreground">
						{t("common.loading", "Loading...")}
					</div>
				) : skills.length === 0 ? (
					<div className="p-4 text-xs text-muted-foreground text-center">
						{t(
							"settings.skills.noSkills",
							"No custom skills installed yet. Click 'Import Skills...' above to import Zeus or external skill sets.",
						)}
					</div>
				) : (
					<div className="divide-y divide-border">
						{skills.map((skill) => (
							<div key={skill.name} className="p-3.5 flex items-start gap-3">
								<div className="p-2 rounded-full bg-secondary text-secondary-foreground mt-0.5 shrink-0">
									<Sparkles className="w-4 h-4 text-amber-500" />
								</div>
								<div className="flex-1 min-w-0">
									<div className="flex items-center gap-2">
										<span className="text-sm font-semibold text-foreground truncate font-mono">
											{skill.name}
										</span>
										{skill.category && (
											<span className="px-2 py-0.5 text-[10px] rounded-full bg-muted text-muted-foreground uppercase font-medium">
												{skill.category}
											</span>
										)}
									</div>
									<p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
										{skill.description ||
											t("settings.skills.noDesc", "No description provided.")}
									</p>
									{skill.triggers && skill.triggers.length > 0 && (
										<div className="flex items-center gap-1.5 mt-2">
											<Terminal className="w-3 h-3 text-muted-foreground" />
											<span className="text-[11px] font-mono text-muted-foreground">
												{skill.triggers.join(", ")}
											</span>
										</div>
									)}
								</div>
							</div>
						))}
					</div>
				)}
			</SettingsSection>
		</>
	);
}
