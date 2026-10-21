import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { components } from "../../api/schema";
import { apiClient, apiErrorMessage } from "../lib/api-client";

export type Team = components["schemas"]["TeamResponse"];
export type TeamRole = components["schemas"]["TeamRoleResponse"];
export type TeamPayload = components["schemas"]["TeamPayload"];

export const teamsQueryKey = ["teams"] as const;

export function useTeams(enabled = true) {
	const query = useQuery({
		queryKey: teamsQueryKey,
		enabled,
		staleTime: 30_000,
		queryFn: async (): Promise<Team[]> => {
			const { data, error } = await apiClient.GET("/api/v1/teams");
			if (error) throw error;
			return (data?.teams ?? []) as Team[];
		},
	});

	return {
		teams: query.data ?? [],
		isLoading: query.isLoading,
		error: query.error ? apiErrorMessage(query.error) : undefined,
	};
}

export function useSaveTeam() {
	const queryClient = useQueryClient();
	const mutation = useMutation({
		mutationFn: async (payload: TeamPayload): Promise<Team> => {
			const { data, error } = await apiClient.POST("/api/v1/teams", {
				body: payload,
			});
			if (error) throw error;
			return data as Team;
		},
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: teamsQueryKey });
		},
	});

	return {
		saveTeam: mutation.mutateAsync,
		isPending: mutation.isPending,
		error: mutation.error ? apiErrorMessage(mutation.error) : undefined,
	};
}

export function useDeleteTeam() {
	const queryClient = useQueryClient();
	const mutation = useMutation({
		mutationFn: async (teamId: string) => {
			const { error } = await apiClient.DELETE("/api/v1/teams/{teamId}", {
				params: { path: { teamId } },
			});
			if (error) throw error;
		},
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: teamsQueryKey });
		},
	});

	return {
		deleteTeam: mutation.mutateAsync,
		isPending: mutation.isPending,
		error: mutation.error ? apiErrorMessage(mutation.error) : undefined,
	};
}
