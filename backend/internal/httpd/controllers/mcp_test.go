package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/aoagents/agent-orchestrator/backend/internal/service/mcp"
)

type mockMCPService struct {
	servers []mcp.ServerConfig
}

func (m *mockMCPService) List(ctx context.Context, harness string) ([]mcp.ServerConfig, error) {
	return m.servers, nil
}

func (m *mockMCPService) Add(ctx context.Context, req mcp.AddServerRequest) (mcp.ServerConfig, error) {
	s := mcp.ServerConfig{
		Name:    req.Name,
		Harness: req.Harness,
		Type:    req.Type,
		Command: req.Command,
		Args:    req.Args,
		URL:     req.URL,
		Env:     req.Env,
		Enabled: true,
	}
	m.servers = append(m.servers, s)
	return s, nil
}

func (m *mockMCPService) Delete(ctx context.Context, harness, name string) error {
	var next []mcp.ServerConfig
	for _, s := range m.servers {
		if s.Name != name {
			next = append(next, s)
		}
	}
	m.servers = next
	return nil
}

func TestMCPServersController(t *testing.T) {
	svc := &mockMCPService{}
	c := &MCPServersController{Svc: svc}
	r := chi.NewRouter()
	c.Register(r)

	// List initially empty
	req := httptest.NewRequest(http.MethodGet, "/mcp/servers", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Add server
	body, _ := json.Marshal(AddMCPServerRequest{
		Name:    "test-srv",
		Harness: "opencode",
		Type:    "local",
		Command: "npx",
		Args:    []string{"-y", "foo"},
	})
	req = httptest.NewRequest(http.MethodPost, "/mcp/servers", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// List again
	req = httptest.NewRequest(http.MethodGet, "/mcp/servers?harness=opencode", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var listResp MCPServersListResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Servers) != 1 || listResp.Servers[0].Name != "test-srv" {
		t.Fatalf("expected 1 server, got %+v", listResp)
	}

	// Delete server
	req = httptest.NewRequest(http.MethodDelete, "/mcp/servers/test-srv?harness=opencode", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
