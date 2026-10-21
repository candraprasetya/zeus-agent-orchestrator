import { Plus, Trash2, Users } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useSkills } from "../../hooks/useSkills";
import { useDeleteTeam, useSaveTeam, useTeams, type Team } from "../../hooks/useTeams";
import { SettingsRow } from "./SettingsRow";
import { SettingsSection } from "./SettingsSection";

export function TeamsSection({ titleHidden }: { titleHidden?: boolean }) {
	const { t } = useTranslation();
	const { teams, isLoading } = useTeams();
	const { saveTeam, isPending: isSaving } = useSaveTeam();
	const { deleteTeam } = useDeleteTeam();
	const { skills } = useSkills();

	const [editingTeam, setEditingTeam] = useState<Team | null>(null);

	const handleCreateNew = () => {
		setEditingTeam({
			id: `team-${Date.now()}`,
			name: "New Squad",
			description: "",
			roles: [
				{
					role: "ui-engineer",
					harness: "mage",
					skills: [],
					instructions: "Handles UI components and layouts.",
				},
			],
		});
	};

	const handleSave = async (team: Team) => {
		await saveTeam(team);
		setEditingTeam(null);
	};

	return (
		<>
			<SettingsSection
				title={t("settings.teams.title", "Squads & Teams")}
				sectionId="teams"
				titleHidden={titleHidden}
				grouped
				action={
					<button
						type="button"
						onClick={handleCreateNew}
						disabled={isSaving}
						className="inline-flex items-center gap-1.5 px-3 py-1 text-xs font-medium bg-primary text-primary-foreground rounded hover:opacity-90 transition"
					>
						<Plus className="w-3.5 h-3.5" />
						{t("settings.teams.createBtn", "New Squad")}
					</button>
				}
			>
				<SettingsRow
					label={t("settings.teams.manageLabel", "Specialized Squads")}
					description={t(
						"settings.teams.manageDesc",
						"Define squads and role-skill mappings (e.g. Zeus iOS/Android squad). Orchestrators use these roles to dispatch tasks from implementation plans.",
					)}
				>
					<span className="text-xs text-muted-foreground font-mono">
						{teams.length} {t("settings.teams.count", "squads")}
					</span>
				</SettingsRow>
			</SettingsSection>

			{editingTeam ? (
				<SettingsSection title={t("settings.teams.editorTitle", "Edit Squad")} grouped>
					<div className="p-4 space-y-4">
						<div>
							<label className="text-xs font-medium text-foreground block mb-1">
								{t("settings.teams.squadName", "Squad Name")}
							</label>
							<input
								type="text"
								value={editingTeam.name}
								onChange={(e) => setEditingTeam({ ...editingTeam, name: e.target.value })}
								className="w-full text-xs px-2.5 py-1.5 rounded border border-border bg-background font-medium"
							/>
						</div>
						<div>
							<label className="text-xs font-medium text-foreground block mb-1">
								{t("settings.teams.squadId", "Squad ID")}
							</label>
							<input
								type="text"
								value={editingTeam.id}
								onChange={(e) => setEditingTeam({ ...editingTeam, id: e.target.value })}
								className="w-full text-xs px-2.5 py-1.5 rounded border border-border bg-background font-mono text-muted-foreground"
							/>
						</div>
						<div>
							<label className="text-xs font-medium text-foreground block mb-1">
								{t("settings.teams.squadDesc", "Description")}
							</label>
							<textarea
								rows={2}
								value={editingTeam.description || ""}
								onChange={(e) => setEditingTeam({ ...editingTeam, description: e.target.value })}
								className="w-full text-xs px-2.5 py-1.5 rounded border border-border bg-background"
							/>
						</div>

						{/* Roles list */}
						<div className="pt-2">
							<div className="flex items-center justify-between mb-2">
								<h4 className="text-xs font-semibold text-foreground">
									{t("settings.teams.rolesList", "Roles & Skills Assignment")}
								</h4>
								<button
									type="button"
									onClick={() =>
										setEditingTeam({
											...editingTeam,
											roles: [
												...editingTeam.roles,
												{
													role: `role-${editingTeam.roles.length + 1}`,
													harness: "mage",
													skills: [],
													instructions: "",
												},
											],
										})
									}
									className="text-[11px] text-primary hover:underline flex items-center gap-1 font-medium"
								>
									<Plus className="w-3 h-3" />
									{t("settings.teams.addRole", "Add Role")}
								</button>
							</div>

							<div className="space-y-3">
								{editingTeam.roles.map((role, idx) => (
									<div key={idx} className="p-3 rounded border border-border bg-card space-y-2">
										<div className="flex items-center gap-2">
											<input
												type="text"
												placeholder="Role identifier (e.g. ui-engineer)"
												value={role.role}
												onChange={(e) => {
													const newRoles = [...editingTeam.roles];
													newRoles[idx] = { ...role, role: e.target.value };
													setEditingTeam({ ...editingTeam, roles: newRoles });
												}}
												className="flex-1 text-xs px-2 py-1 rounded border border-border bg-background font-mono font-medium"
											/>
											<select
												value={role.harness || "mage"}
												onChange={(e) => {
													const newRoles = [...editingTeam.roles];
													newRoles[idx] = { ...role, harness: e.target.value };
													setEditingTeam({ ...editingTeam, roles: newRoles });
												}}
												className="text-xs px-2 py-1 rounded border border-border bg-background font-medium"
											>
												<option value="mage">Mage</option>
												<option value="opencode">OpenCode</option>
												<option value="claude-code">Claude Code</option>
												<option value="codex">Codex</option>
											</select>
											<button
												type="button"
												onClick={() => {
													const newRoles = editingTeam.roles.filter((_, i) => i !== idx);
													setEditingTeam({ ...editingTeam, roles: newRoles });
												}}
												className="text-muted-foreground hover:text-destructive p-1"
											>
												<Trash2 className="w-3.5 h-3.5" />
											</button>
										</div>

										<div>
											<label className="text-[10px] text-muted-foreground uppercase font-semibold block mb-1">
												Assigned Skills
											</label>
											<div className="flex flex-wrap gap-1.5 max-h-24 overflow-y-auto p-1.5 rounded border border-border/50 bg-background/50">
												{skills.length === 0 ? (
													<span className="text-[11px] text-muted-foreground">
														No skills available. Import skills first in the Skills tab.
													</span>
												) : (
													skills.map((s) => {
														const selected = role.skills.includes(s.name);
														return (
															<button
																key={s.name}
																type="button"
																onClick={() => {
																	const newRoles = [...editingTeam.roles];
																	const nextSkills = selected
																		? role.skills.filter((name) => name !== s.name)
																		: [...role.skills, s.name];
																	newRoles[idx] = { ...role, skills: nextSkills };
																	setEditingTeam({ ...editingTeam, roles: newRoles });
																}}
																className={`text-[10px] px-2 py-0.5 rounded font-mono transition ${
																	selected
																		? "bg-primary text-primary-foreground font-semibold"
																		: "bg-muted text-muted-foreground hover:bg-muted/80"
																}`}
															>
																{s.name}
															</button>
														);
													})
												)}
											</div>
										</div>
									</div>
								))}
							</div>
						</div>

						<div className="flex items-center justify-end gap-2 pt-3 border-t border-border">
							<button
								type="button"
								onClick={() => setEditingTeam(null)}
								className="px-3 py-1.5 text-xs text-muted-foreground hover:text-foreground"
							>
								{t("common.cancel", "Cancel")}
							</button>
							<button
								type="button"
								onClick={() => void handleSave(editingTeam)}
								disabled={isSaving}
								className="px-3 py-1.5 text-xs font-medium bg-primary text-primary-foreground rounded hover:opacity-90"
							>
								{t("common.save", "Save Squad")}
							</button>
						</div>
					</div>
				</SettingsSection>
			) : (
				<SettingsSection title={t("settings.teams.squadsList", "Configured Squads")} grouped>
					{isLoading ? (
						<div className="p-4 text-xs text-muted-foreground">{t("common.loading", "Loading...")}</div>
					) : teams.length === 0 ? (
						<div className="p-4 text-xs text-muted-foreground text-center">
							{t("settings.teams.noTeams", "No squads configured yet. Click 'New Squad' to create one.")}
						</div>
					) : (
						<div className="divide-y divide-border">
							{teams.map((team) => (
								<div key={team.id} className="p-3.5 flex items-start gap-3">
									<div className="p-2 rounded-md bg-secondary text-secondary-foreground mt-0.5">
										<Users className="w-4 h-4 text-blue-500" />
									</div>
									<div className="flex-1 min-w-0">
										<div className="flex items-center gap-2">
											<span className="text-sm font-semibold text-foreground truncate">
												{team.name}
											</span>
											<span className="text-[11px] font-mono text-muted-foreground">
												({team.id})
											</span>
										</div>
										{team.description && (
											<p className="text-xs text-muted-foreground mt-0.5 line-clamp-1">
												{team.description}
											</p>
										)}
										<div className="mt-2 flex flex-wrap gap-2">
											{team.roles.map((r) => (
												<div
													key={r.role}
													className="text-[11px] px-2 py-1 rounded bg-muted/70 text-foreground border border-border/40 font-mono"
												>
													<span className="font-semibold">{r.role}</span>
													<span className="text-muted-foreground ml-1.5">
														({r.harness} · {r.skills.length} skills)
													</span>
												</div>
											))}
										</div>
									</div>
									<div className="flex items-center gap-1">
										<button
											type="button"
											onClick={() => setEditingTeam(team)}
											className="px-2.5 py-1 text-xs text-primary hover:bg-primary/10 rounded font-medium"
										>
											{t("common.edit", "Edit")}
										</button>
										<button
											type="button"
											onClick={() => void deleteTeam(team.id)}
											className="p-1 text-muted-foreground hover:text-destructive"
										>
											<Trash2 className="w-3.5 h-3.5" />
										</button>
									</div>
								</div>
							))}
						</div>
					)}
				</SettingsSection>
			)}
		</>
	);
}
