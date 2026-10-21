package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apispec"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/envelope"
	"github.com/aoagents/agent-orchestrator/backend/internal/service/teams"
)

// TeamsService defines the controller contract for managing squads and roles.
type TeamsService interface {
	List(ctx context.Context) ([]teams.Team, error)
	Get(ctx context.Context, id string) (teams.Team, error)
	Save(ctx context.Context, team teams.Team) (teams.Team, error)
	Delete(ctx context.Context, id string) error
	GetTeamPrompt(ctx context.Context, teamID string) (string, error)
}

// TeamsController exposes REST endpoints for Teams and Squads.
type TeamsController struct {
	Svc TeamsService
}

// Register mounts the /api/v1/teams routes.
func (c *TeamsController) Register(r chi.Router) {
	r.Get("/teams", c.list)
	r.Post("/teams", c.save)
	r.Get("/teams/{teamId}", c.get)
	r.Put("/teams/{teamId}", c.save)
	r.Delete("/teams/{teamId}", c.delete)
	r.Get("/teams/{teamId}/prompt", c.prompt)
}

func (c *TeamsController) list(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "GET", "/api/v1/teams")
		return
	}
	list, err := c.Svc.List(r.Context())
	if err != nil {
		envelope.WriteAPIError(w, r, http.StatusInternalServerError, "internal_error", "TEAMS_LIST_FAILED", err.Error(), nil)
		return
	}
	resp := TeamsListResponse{
		Teams: make([]TeamResponse, 0, len(list)),
	}
	for _, t := range list {
		resp.Teams = append(resp.Teams, toTeamResponse(t))
	}
	envelope.WriteJSON(w, http.StatusOK, resp)
}

func (c *TeamsController) get(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "GET", "/api/v1/teams/{teamId}")
		return
	}
	id := chi.URLParam(r, "teamId")
	team, err := c.Svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, teams.ErrTeamNotFound) {
			envelope.WriteAPIError(w, r, http.StatusNotFound, "not_found", "TEAM_NOT_FOUND", err.Error(), nil)
			return
		}
		envelope.WriteAPIError(w, r, http.StatusInternalServerError, "internal_error", "TEAM_GET_FAILED", err.Error(), nil)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, toTeamResponse(team))
}

func (c *TeamsController) save(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "POST", "/api/v1/teams")
		return
	}
	var req TeamPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "INVALID_JSON", "invalid JSON payload", nil)
		return
	}
	urlID := chi.URLParam(r, "teamId")
	if urlID != "" {
		req.ID = urlID
	}
	saved, err := c.Svc.Save(r.Context(), fromTeamPayload(req))
	if err != nil {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "SAVE_TEAM_FAILED", err.Error(), nil)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, toTeamResponse(saved))
}

func (c *TeamsController) delete(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "DELETE", "/api/v1/teams/{teamId}")
		return
	}
	id := chi.URLParam(r, "teamId")
	if err := c.Svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, teams.ErrTeamNotFound) {
			envelope.WriteAPIError(w, r, http.StatusNotFound, "not_found", "TEAM_NOT_FOUND", err.Error(), nil)
			return
		}
		envelope.WriteAPIError(w, r, http.StatusInternalServerError, "internal_error", "DELETE_TEAM_FAILED", err.Error(), nil)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, map[string]any{"status": "deleted", "id": id})
}

func (c *TeamsController) prompt(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "GET", "/api/v1/teams/{teamId}/prompt")
		return
	}
	id := chi.URLParam(r, "teamId")
	prompt, err := c.Svc.GetTeamPrompt(r.Context(), id)
	if err != nil {
		if errors.Is(err, teams.ErrTeamNotFound) {
			envelope.WriteAPIError(w, r, http.StatusNotFound, "not_found", "TEAM_NOT_FOUND", err.Error(), nil)
			return
		}
		envelope.WriteAPIError(w, r, http.StatusInternalServerError, "internal_error", "GET_PROMPT_FAILED", err.Error(), nil)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, TeamPromptResponse{Prompt: prompt})
}

func toTeamResponse(t teams.Team) TeamResponse {
	roles := make([]TeamRoleResponse, 0, len(t.Roles))
	for _, r := range t.Roles {
		roles = append(roles, TeamRoleResponse{
			Role:         r.Role,
			Harness:      r.Harness,
			Model:        r.Model,
			Skills:       r.Skills,
			Instructions: r.Instructions,
		})
	}
	return TeamResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Roles:       roles,
	}
}

func fromTeamPayload(p TeamPayload) teams.Team {
	roles := make([]teams.RoleMember, 0, len(p.Roles))
	for _, r := range p.Roles {
		roles = append(roles, teams.RoleMember{
			Role:         r.Role,
			Harness:      r.Harness,
			Model:        r.Model,
			Skills:       r.Skills,
			Instructions: r.Instructions,
		})
	}
	return teams.Team{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Roles:       roles,
	}
}
