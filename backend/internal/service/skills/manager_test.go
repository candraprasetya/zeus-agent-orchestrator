package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestManagerListAndImport(t *testing.T) {
	dataDir := t.TempDir()
	mgr := New(dataDir)

	// Initially empty
	list, err := mgr.List(context.Background())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 skills, got %d", len(list))
	}

	// Create a dummy skill in another folder
	sourceDir := t.TempDir()
	skillADir := filepath.Join(sourceDir, "mobile", "zeus-android")
	if err := os.MkdirAll(skillADir, 0o750); err != nil {
		t.Fatal(err)
	}
	skillAContent := `---
name: zeus-android
description: Android squad skill for testing
triggers:
  - "/zeus-android"
---
# Zeus Android
Sample skill documentation.
`
	if err := os.WriteFile(filepath.Join(skillADir, "SKILL.md"), []byte(skillAContent), 0o600); err != nil {
		t.Fatal(err)
	}

	// Import skills from sourceDir
	imported, err := mgr.Import(context.Background(), sourceDir)
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}
	if len(imported) != 1 {
		t.Fatalf("expected 1 imported skill, got %d", len(imported))
	}
	if imported[0].Name != "zeus-android" {
		t.Errorf("imported name = %q, want zeus-android", imported[0].Name)
	}
	if imported[0].Description != "Android squad skill for testing" {
		t.Errorf("imported description = %q", imported[0].Description)
	}

	// Verify List now returns the imported skill
	skills, err := mgr.List(context.Background())
	if err != nil {
		t.Fatalf("List error after import: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill in list, got %d", len(skills))
	}

	// Verify Get
	skill, err := mgr.Get(context.Background(), "zeus-android")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if skill.Name != "zeus-android" {
		t.Errorf("Get name = %q, want zeus-android", skill.Name)
	}
}
