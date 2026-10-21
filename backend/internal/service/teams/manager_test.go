package teams

import (
	"context"
	"strings"
	"testing"
)

func TestTeamsManager(t *testing.T) {
	dataDir := t.TempDir()
	mgr := New(dataDir)

	// List should return default teams
	list, err := mgr.List(context.Background())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) < 2 {
		t.Fatalf("expected at least 2 default teams, got %d", len(list))
	}

	// Get zeus-team-ios
	team, err := mgr.Get(context.Background(), "zeus-team-ios")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if team.Name != "Zeus iOS Squad" {
		t.Errorf("expected Zeus iOS Squad, got %s", team.Name)
	}

	// Get prompt
	prompt, err := mgr.GetTeamPrompt(context.Background(), "zeus-team-ios")
	if err != nil {
		t.Fatalf("GetTeamPrompt error: %v", err)
	}
	if !strings.Contains(prompt, "ui-engineer") || !strings.Contains(prompt, "zeus-ios-presentation") {
		t.Errorf("prompt missing expected role info:\n%s", prompt)
	}

	// Save custom team
	custom := Team{
		ID:   "custom-web",
		Name: "Custom Web Squad",
		Roles: []RoleMember{
			{Role: "frontend", Harness: "mage", Skills: []string{"react-skill"}},
		},
	}
	if _, err := mgr.Save(context.Background(), custom); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	saved, err := mgr.Get(context.Background(), "custom-web")
	if err != nil || saved.Name != "Custom Web Squad" {
		t.Fatalf("expected custom-web saved, got %v, err=%v", saved, err)
	}
}
