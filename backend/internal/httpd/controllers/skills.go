package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apispec"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/envelope"
	"github.com/aoagents/agent-orchestrator/backend/internal/service/skills"
)

// SkillsService defines the operations exposed by the skills management controller.
type SkillsService interface {
	List(ctx context.Context) ([]skills.Skill, error)
	Get(ctx context.Context, name string) (skills.Skill, error)
	Import(ctx context.Context, srcPath string) ([]skills.Skill, error)
}

// SkillsController serves REST endpoints for skills management.
type SkillsController struct {
	Svc SkillsService
}

// Register mounts the skills endpoints under /api/v1/skills.
func (c *SkillsController) Register(r chi.Router) {
	r.Get("/skills", c.list)
	r.Post("/skills/import", c.importSkills)
}

func (c *SkillsController) list(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "GET", "/api/v1/skills")
		return
	}

	items, err := c.Svc.List(r.Context())
	if err != nil {
		envelope.WriteAPIError(w, r, http.StatusInternalServerError, "internal_error", "SKILLS_LIST_FAILED", err.Error(), nil)
		return
	}

	resp := SkillsListResponse{
		Skills: make([]SkillResponse, 0, len(items)),
	}
	for _, item := range items {
		resp.Skills = append(resp.Skills, SkillResponse{
			Name:        item.Name,
			Description: item.Description,
			Path:        item.Path,
			Category:    item.Category,
			Triggers:    item.Triggers,
			IsManaged:   item.IsManaged,
		})
	}

	envelope.WriteJSON(w, http.StatusOK, resp)
}

func (c *SkillsController) importSkills(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "POST", "/api/v1/skills/import")
		return
	}

	var req ImportSkillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "INVALID_JSON", "invalid JSON payload", nil)
		return
	}

	if req.SourcePath == "" {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "SOURCE_PATH_REQUIRED", "sourcePath is required", nil)
		return
	}

	imported, err := c.Svc.Import(r.Context(), req.SourcePath)
	if err != nil {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "IMPORT_FAILED", err.Error(), nil)
		return
	}

	resp := SkillsListResponse{
		Skills: make([]SkillResponse, 0, len(imported)),
	}
	for _, item := range imported {
		resp.Skills = append(resp.Skills, SkillResponse{
			Name:        item.Name,
			Description: item.Description,
			Path:        item.Path,
			Category:    item.Category,
			Triggers:    item.Triggers,
			IsManaged:   item.IsManaged,
		})
	}

	envelope.WriteJSON(w, http.StatusOK, resp)
}
