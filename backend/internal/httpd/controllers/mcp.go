package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apispec"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/envelope"
	"github.com/aoagents/agent-orchestrator/backend/internal/service/mcp"
)

// MCPService defines the operations exposed by the MCP management controller.
type MCPService interface {
	List(ctx context.Context, harness string) ([]mcp.ServerConfig, error)
	Add(ctx context.Context, req mcp.AddServerRequest) (mcp.ServerConfig, error)
	Delete(ctx context.Context, harness, name string) error
}

// MCPServersController serves REST endpoints for MCP servers management.
type MCPServersController struct {
	Svc MCPService
}

// Register mounts the MCP endpoints under /api/v1/mcp.
func (c *MCPServersController) Register(r chi.Router) {
	r.Get("/mcp/servers", c.list)
	r.Post("/mcp/servers", c.add)
	r.Delete("/mcp/servers/{name}", c.delete)
}

func (c *MCPServersController) list(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "GET", "/api/v1/mcp/servers")
		return
	}

	harness := r.URL.Query().Get("harness")
	items, err := c.Svc.List(r.Context(), harness)
	if err != nil {
		envelope.WriteAPIError(w, r, http.StatusInternalServerError, "internal_error", "MCP_LIST_FAILED", err.Error(), nil)
		return
	}

	resp := MCPServersListResponse{
		Servers: make([]MCPServerResponse, 0, len(items)),
	}
	for _, item := range items {
		resp.Servers = append(resp.Servers, MCPServerResponse{
			Name:       item.Name,
			Harness:    item.Harness,
			Type:       item.Type,
			Command:    item.Command,
			Args:       item.Args,
			URL:        item.URL,
			Env:        item.Env,
			ConfigPath: item.ConfigPath,
			Enabled:    item.Enabled,
		})
	}

	envelope.WriteJSON(w, http.StatusOK, resp)
}

func (c *MCPServersController) add(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "POST", "/api/v1/mcp/servers")
		return
	}

	var req AddMCPServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "INVALID_JSON", "invalid JSON payload", nil)
		return
	}

	if req.Name == "" {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "NAME_REQUIRED", "server name is required", nil)
		return
	}

	added, err := c.Svc.Add(r.Context(), mcp.AddServerRequest{
		Name:    req.Name,
		Harness: req.Harness,
		Type:    req.Type,
		Command: req.Command,
		Args:    req.Args,
		URL:     req.URL,
		Env:     req.Env,
	})
	if err != nil {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "ADD_FAILED", err.Error(), nil)
		return
	}

	resp := MCPServerResponse{
		Name:       added.Name,
		Harness:    added.Harness,
		Type:       added.Type,
		Command:    added.Command,
		Args:       added.Args,
		URL:        added.URL,
		Env:        added.Env,
		ConfigPath: added.ConfigPath,
		Enabled:    added.Enabled,
	}

	envelope.WriteJSON(w, http.StatusOK, resp)
}

func (c *MCPServersController) delete(w http.ResponseWriter, r *http.Request) {
	if c.Svc == nil {
		apispec.NotImplemented(w, r, "DELETE", "/api/v1/mcp/servers/{name}")
		return
	}

	name := chi.URLParam(r, "name")
	if name == "" {
		envelope.WriteAPIError(w, r, http.StatusBadRequest, "bad_request", "NAME_REQUIRED", "server name is required", nil)
		return
	}

	harness := r.URL.Query().Get("harness")
	if err := c.Svc.Delete(r.Context(), harness, name); err != nil {
		envelope.WriteAPIError(w, r, http.StatusInternalServerError, "internal_error", "DELETE_FAILED", err.Error(), nil)
		return
	}

	envelope.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
