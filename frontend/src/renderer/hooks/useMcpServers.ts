import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { components } from "../../api/schema";
import { apiClient, apiErrorMessage } from "../lib/api-client";

export type MCPServer = components["schemas"]["MCPServerResponse"];
export type AddMCPServerPayload = components["schemas"]["AddMCPServerRequest"];

export const mcpServersQueryKey = (harness?: string) => ["mcpServers", harness ?? "all"] as const;

export function useMcpServers(harness = "opencode") {
	const query = useQuery({
		queryKey: mcpServersQueryKey(harness),
		staleTime: 10_000,
		queryFn: async (): Promise<MCPServer[]> => {
			const { data, error } = await apiClient.GET("/api/v1/mcp/servers", {
				params: {
					query: {
						harness: harness || undefined,
					},
				},
			});
			if (error) throw error;
			return (data?.servers ?? []) as MCPServer[];
		},
	});

	return {
		servers: query.data ?? [],
		isLoading: query.isLoading,
		error: query.error ? apiErrorMessage(query.error) : undefined,
		refetch: query.refetch,
	};
}

export function useAddMcpServer() {
	const queryClient = useQueryClient();
	const mutation = useMutation({
		mutationFn: async (payload: AddMCPServerPayload): Promise<MCPServer> => {
			const { data, error } = await apiClient.POST("/api/v1/mcp/servers", {
				body: payload,
			});
			if (error) throw error;
			return data as MCPServer;
		},
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["mcpServers"] });
		},
	});

	return {
		addServer: mutation.mutateAsync,
		isPending: mutation.isPending,
		error: mutation.error ? apiErrorMessage(mutation.error) : undefined,
	};
}

export function useDeleteMcpServer() {
	const queryClient = useQueryClient();
	const mutation = useMutation({
		mutationFn: async ({ name, harness }: { name: string; harness: string }): Promise<void> => {
			const { error } = await apiClient.DELETE("/api/v1/mcp/servers/{name}", {
				params: {
					path: { name },
					query: { harness },
				},
			});
			if (error) throw error;
		},
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["mcpServers"] });
		},
	});

	return {
		deleteServer: mutation.mutateAsync,
		isPending: mutation.isPending,
		error: mutation.error ? apiErrorMessage(mutation.error) : undefined,
	};
}
