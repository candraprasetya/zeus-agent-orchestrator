import { Globe, Loader2, Plug, Plus, Server, Terminal, Trash2, X } from "@/icons/hugeicons-adapter";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useAddMcpServer, useDeleteMcpServer, useMcpServers } from "../../hooks/useMcpServers";
import { SettingsRow } from "./SettingsRow";
import { SettingsSection } from "./SettingsSection";

export function McpServersSection({ titleHidden }: { titleHidden?: boolean }) {
	const { t } = useTranslation();
	const [selectedHarness, setSelectedHarness] = useState<string>("opencode");
	const [showAddForm, setShowAddForm] = useState(false);

	// Form fields
	const [name, setName] = useState("");
	const [targetHarness, setTargetHarness] = useState("opencode");
	const [transportType, setTransportType] = useState<"local" | "remote">("local");
	const [command, setCommand] = useState("");
	const [argsText, setArgsText] = useState("");
	const [url, setUrl] = useState("");
	const [envText, setEnvText] = useState("");
	const [actionError, setActionError] = useState<string | null>(null);

	const { servers, isLoading, refetch } = useMcpServers(selectedHarness);
	const { addServer, isPending: isAdding } = useAddMcpServer();
	const { deleteServer, isPending: isDeleting } = useDeleteMcpServer();

	const handleAdd = async (e: React.FormEvent) => {
		e.preventDefault();
		setActionError(null);

		const trimmedName = name.trim();
		if (!trimmedName) {
			setActionError("Server name is required");
			return;
		}

		if (transportType === "local" && !command.trim()) {
			setActionError("Command is required for local server");
			return;
		}

		if (transportType === "remote" && !url.trim()) {
			setActionError("URL is required for remote server");
			return;
		}

		const parsedArgs = argsText
			.split(/\s+/)
			.map((a) => a.trim())
			.filter(Boolean);

		const parsedEnv: Record<string, string> = {};
		if (envText.trim()) {
			for (const line of envText.split("\n")) {
				const trimmed = line.trim();
				if (!trimmed || trimmed.startsWith("#")) continue;
				const eqIdx = trimmed.indexOf("=");
				if (eqIdx > 0) {
					const key = trimmed.slice(0, eqIdx).trim();
					const val = trimmed.slice(eqIdx + 1).trim();
					parsedEnv[key] = val;
				}
			}
		}

		try {
			await addServer({
				name: trimmedName,
				harness: targetHarness,
				type: transportType,
				command: command.trim() || undefined,
				args: parsedArgs.length > 0 ? parsedArgs : undefined,
				url: url.trim() || undefined,
				env: Object.keys(parsedEnv).length > 0 ? parsedEnv : undefined,
			});

			// Reset form
			setName("");
			setCommand("");
			setArgsText("");
			setUrl("");
			setEnvText("");
			setShowAddForm(false);
			void refetch();
		} catch (err) {
			setActionError(err instanceof Error ? err.message : String(err));
		}
	};

	const handleDelete = async (serverName: string, harness: string) => {
		if (!confirm(`Are you sure you want to remove MCP server "${serverName}"?`)) return;
		try {
			await deleteServer({ name: serverName, harness });
			void refetch();
		} catch (err) {
			setActionError(err instanceof Error ? err.message : String(err));
		}
	};

	const harnessOptions = [
		{ id: "opencode", label: "OpenCode" },
		{ id: "mage", label: "Mage" },
		{ id: "agy", label: "AGY (Antigravity)" },
		{ id: "claude", label: "Claude Code" },
		{ id: "all", label: "All Harnesses" },
	];

	return (
		<>
			<SettingsSection
				title={t("settings.mcp.title", "Model Context Protocol (MCP)")}
				sectionId="mcp"
				titleHidden={titleHidden}
				grouped
				action={
					<button
						type="button"
						onClick={() => setShowAddForm((prev) => !prev)}
						className="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-medium bg-primary text-primary-foreground rounded-full hover:opacity-90 transition shadow-sm"
					>
						{showAddForm ? <X className="w-3.5 h-3.5" /> : <Plus className="w-3.5 h-3.5" />}
						{showAddForm ? t("common.cancel", "Cancel") : t("settings.mcp.addBtn", "Add Server...")}
					</button>
				}
			>
				<SettingsRow
					label={t("settings.mcp.harnessFilter", "Target Harness")}
					description={t(
						"settings.mcp.filterDesc",
						"View and configure MCP servers connected to your AI agent harnesses.",
					)}
				>
					<div className="flex items-center gap-1 bg-secondary/50 p-1 rounded-lg">
						{harnessOptions.map((opt) => (
							<button
								key={opt.id}
								type="button"
								onClick={() => setSelectedHarness(opt.id)}
								className={`px-2.5 py-1 text-xs rounded-md font-medium transition ${
									selectedHarness === opt.id
										? "bg-background text-foreground shadow-sm"
										: "text-muted-foreground hover:text-foreground"
								}`}
							>
								{opt.label}
							</button>
						))}
					</div>
				</SettingsRow>
			</SettingsSection>

			{/* Add MCP Server Form */}
			{showAddForm && (
				<div className="bg-card border border-border rounded-xl p-4 shadow-sm mb-4">
					<div className="flex items-center justify-between pb-3 border-b border-border mb-3">
						<div className="flex items-center gap-2">
							<Plug className="w-4 h-4 text-primary" />
							<h3 className="text-sm font-semibold text-foreground">
								{t("settings.mcp.addServerTitle", "Register New MCP Server")}
							</h3>
						</div>
						<button
							type="button"
							onClick={() => setShowAddForm(false)}
							className="text-muted-foreground hover:text-foreground p-1"
						>
							<X className="w-4 h-4" />
						</button>
					</div>

					{actionError && (
						<div className="mb-3 p-2.5 text-xs rounded-md bg-destructive/10 text-destructive border border-destructive/20">
							{actionError}
						</div>
					)}

					<form onSubmit={handleAdd} className="space-y-3">
						<div className="grid grid-cols-2 gap-3">
							<div>
								<label className="block text-xs font-medium text-foreground mb-1">
									{t("settings.mcp.serverName", "Server Name")} *
								</label>
								<input
									type="text"
									value={name}
									onChange={(e) => setName(e.target.value)}
									placeholder="e.g. memory-server"
									className="w-full px-3 py-1.5 text-xs bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-1 focus:ring-primary font-mono"
									required
								/>
							</div>

							<div>
								<label className="block text-xs font-medium text-foreground mb-1">
									{t("settings.mcp.targetHarness", "Harness")}
								</label>
								<select
									value={targetHarness}
									onChange={(e) => setTargetHarness(e.target.value)}
									className="w-full px-3 py-1.5 text-xs bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
								>
									<option value="opencode">OpenCode (~/.config/opencode/opencode.json)</option>
									<option value="mage">Mage (~/.mage/mage.json)</option>
									<option value="agy">AGY (~/.gemini/config/mcp_config.json)</option>
									<option value="claude">Claude Code (~/.claude.json)</option>
								</select>
							</div>
						</div>

						<div>
							<label className="block text-xs font-medium text-foreground mb-1">
								{t("settings.mcp.transportType", "Transport Type")}
							</label>
							<div className="flex items-center gap-3">
								<label className="inline-flex items-center gap-1.5 text-xs cursor-pointer">
									<input
										type="radio"
										name="transport"
										value="local"
										checked={transportType === "local"}
										onChange={() => setTransportType("local")}
										className="text-primary"
									/>
									Local (Stdio Command)
								</label>
								<label className="inline-flex items-center gap-1.5 text-xs cursor-pointer">
									<input
										type="radio"
										name="transport"
										value="remote"
										checked={transportType === "remote"}
										onChange={() => setTransportType("remote")}
										className="text-primary"
									/>
									Remote (HTTP / SSE URL)
								</label>
							</div>
						</div>

						{transportType === "local" ? (
							<>
								<div className="grid grid-cols-2 gap-3">
									<div>
										<label className="block text-xs font-medium text-foreground mb-1">
											{t("settings.mcp.command", "Command")} *
										</label>
										<input
											type="text"
											value={command}
											onChange={(e) => setCommand(e.target.value)}
											placeholder="e.g. npx, uvx, node, docker"
											className="w-full px-3 py-1.5 text-xs bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-1 focus:ring-primary font-mono"
											required={transportType === "local"}
										/>
									</div>
									<div>
										<label className="block text-xs font-medium text-foreground mb-1">
											{t("settings.mcp.args", "Arguments")}
										</label>
										<input
											type="text"
											value={argsText}
											onChange={(e) => setArgsText(e.target.value)}
											placeholder="e.g. -y @modelcontextprotocol/server-memory"
											className="w-full px-3 py-1.5 text-xs bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-1 focus:ring-primary font-mono"
										/>
									</div>
								</div>

								<div>
									<label className="block text-xs font-medium text-foreground mb-1">
										{t("settings.mcp.envVars", "Environment Variables (KEY=VALUE, one per line)")}
									</label>
									<textarea
										rows={2}
										value={envText}
										onChange={(e) => setEnvText(e.target.value)}
										placeholder="API_KEY=your_token_here&#10;PORT=3000"
										className="w-full px-3 py-1.5 text-xs bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-1 focus:ring-primary font-mono"
									/>
								</div>
							</>
						) : (
							<div>
								<label className="block text-xs font-medium text-foreground mb-1">
									{t("settings.mcp.remoteUrl", "Endpoint URL")} *
								</label>
								<input
									type="url"
									value={url}
									onChange={(e) => setUrl(e.target.value)}
									placeholder="https://mcp.example.com/mcp"
									className="w-full px-3 py-1.5 text-xs bg-background border border-border rounded-md text-foreground focus:outline-none focus:ring-1 focus:ring-primary font-mono"
									required={transportType === "remote"}
								/>
							</div>
						)}

						<div className="flex justify-end gap-2 pt-2">
							<button
								type="button"
								onClick={() => setShowAddForm(false)}
								className="px-3 py-1.5 text-xs rounded-md text-muted-foreground hover:text-foreground"
							>
								{t("common.cancel", "Cancel")}
							</button>
							<button
								type="submit"
								disabled={isAdding}
								className="inline-flex items-center gap-1.5 px-4 py-1.5 text-xs font-medium bg-primary text-primary-foreground rounded-md hover:opacity-90 disabled:opacity-50 transition"
							>
								{isAdding && <Loader2 className="w-3 h-3 animate-spin" />}
								{t("settings.mcp.saveBtn", "Save Server")}
							</button>
						</div>
					</form>
				</div>
			)}

			{/* Server List */}
			<SettingsSection
				title={t("settings.mcp.configuredTitle", "Configured MCP Servers")}
				grouped
			>
				{isLoading ? (
					<div className="p-4 text-xs text-muted-foreground flex items-center gap-2">
						<Loader2 className="w-4 h-4 animate-spin" />
						{t("common.loading", "Loading configured MCP servers...")}
					</div>
				) : servers.length === 0 ? (
					<div className="p-5 text-center text-muted-foreground">
						<Server className="w-6 h-6 mx-auto mb-2 opacity-40" />
						<p className="text-xs">
							{t("settings.mcp.emptyState", "No MCP servers configured for this harness.")}
						</p>
						<p className="text-[11px] mt-1 opacity-70">
							Click &quot;Add Server...&quot; above to connect external tools, database, or API integrations.
						</p>
					</div>
				) : (
					<div className="divide-y divide-border">
						{servers.map((server) => (
							<div key={`${server.harness}-${server.name}`} className="p-3.5 flex items-start gap-3">
								<div className="p-2 rounded-full bg-secondary text-secondary-foreground mt-0.5 shrink-0">
									{server.type === "remote" ? (
										<Globe className="w-4 h-4 text-sky-500" />
									) : (
										<Terminal className="w-4 h-4 text-emerald-500" />
									)}
								</div>

								<div className="flex-1 min-w-0">
									<div className="flex items-center gap-2 flex-wrap">
										<span className="text-sm font-semibold text-foreground font-mono">
											{server.name}
										</span>
										<span className="px-2 py-0.5 text-[10px] rounded-full bg-primary/10 text-primary font-medium uppercase">
											{server.harness}
										</span>
										<span className="px-2 py-0.5 text-[10px] rounded-full bg-muted text-muted-foreground font-mono">
											{server.type === "remote" ? "Remote URL" : "Stdio"}
										</span>
									</div>

									<div className="mt-1 text-xs font-mono text-muted-foreground bg-secondary/30 rounded px-2 py-1 truncate">
										{server.type === "remote"
											? server.url
											: `${server.command || ""} ${(server.args || []).join(" ")}`}
									</div>

									{server.configPath && (
										<p className="text-[10px] text-muted-foreground/70 mt-1 truncate">
											Config: {server.configPath}
										</p>
									)}
								</div>

								<button
									type="button"
									onClick={() => void handleDelete(server.name, server.harness)}
									disabled={isDeleting}
									title="Remove server"
									className="p-1.5 text-muted-foreground hover:text-destructive rounded-md hover:bg-destructive/10 transition shrink-0"
								>
									<Trash2 className="w-4 h-4" />
								</button>
							</div>
						))}
					</div>
				)}
			</SettingsSection>
		</>
	);
}
