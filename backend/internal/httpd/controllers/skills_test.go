package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/aoagents/agent-orchestrator/backend/internal/service/skills"
)

type fakeSkillsService struct {
	items []skills.Skill
}

func (f *fakeSkillsService) List(ctx context.Context) ([]skills.Skill, error) {
	return f.items, nil
}

func (f *fakeSkillsService) Get(ctx context.Context, name string) (skills.Skill, error) {
	for _, item := range f.items {
		if item.Name == name {
			return item, nil
		}
	}
	return skills.Skill{}, skills.ErrSkillNotFound
}

func (f *fakeSkillsService) Import(ctx context.Context, srcPath string) ([]skills.Skill, error) {
	imported := skills.Skill{
		Name:        "imported-skill",
		Description: "imported from " + srcPath,
		IsManaged:   true,
	}
	f.items = append(f.items, imported)
	return []skills.Skill{imported}, nil
}

func TestSkillsController(t *testing.T) {
	svc := &fakeSkillsService{
		items: []skills.Skill{
			{Name: "using-ao", Description: "Core AO skill", IsManaged: true},
		},
	}
	c := &SkillsController{Svc: svc}
	r := chi.NewRouter()
	c.Register(r)

	// Test GET /skills
	req := httptest.NewRequest(http.MethodGet, "/skills", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /skills = %d, want 200", rec.Code)
	}

	var listResp SkillsListResponse
	if err := json.NewDecoder(rec.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(listResp.Skills) != 1 || listResp.Skills[0].Name != "using-ao" {
		t.Errorf("skills = %+v, want using-ao", listResp.Skills)
	}

	// Test POST /skills/import
	importBody := bytes.NewBufferString(`{"sourcePath":"/tmp/test"}`)
	req = httptest.NewRequest(http.MethodPost, "/skills/import", importBody)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /skills/import = %d, want 200", rec.Code)
	}

	var importResp SkillsListResponse
	if err := json.NewDecoder(rec.Body).Decode(&importResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(importResp.Skills) != 1 || importResp.Skills[0].Name != "imported-skill" {
		t.Errorf("imported = %+v, want imported-skill", importResp.Skills)
	}
}
