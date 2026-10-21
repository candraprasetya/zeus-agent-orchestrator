import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { components } from "../../api/schema";
import { apiClient, apiErrorMessage } from "../lib/api-client";

export type Skill = components["schemas"]["SkillResponse"];

export const skillsQueryKey = ["skills"] as const;

export function useSkills(enabled = true) {
	const query = useQuery({
		queryKey: skillsQueryKey,
		enabled,
		staleTime: 30_000,
		queryFn: async (): Promise<Skill[]> => {
			const { data, error } = await apiClient.GET("/api/v1/skills");
			if (error) throw error;
			return (data?.skills ?? []) as Skill[];
		},
	});

	return {
		skills: query.data ?? [],
		isLoading: query.isLoading,
		error: query.error ? apiErrorMessage(query.error) : undefined,
	};
}

export function useImportSkills() {
	const queryClient = useQueryClient();
	const mutation = useMutation({
		mutationFn: async (sourcePath: string): Promise<Skill[]> => {
			const { data, error } = await apiClient.POST("/api/v1/skills/import", {
				body: { sourcePath },
			});
			if (error) throw error;
			return (data?.skills ?? []) as Skill[];
		},
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: skillsQueryKey });
		},
	});

	return {
		importSkills: mutation.mutateAsync,
		isPending: mutation.isPending,
		error: mutation.error ? apiErrorMessage(mutation.error) : undefined,
	};
}
