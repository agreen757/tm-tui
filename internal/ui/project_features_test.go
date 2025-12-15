package ui

import (
	"strings"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
	"github.com/agreen757/tm-tui/internal/projects"
)

func TestProjectListItemIncludesTag(t *testing.T) {
	meta := &projects.Metadata{
		Name: "Demo",
		Path: "/tmp/demo",
		Tags: []string{"demo"},
	}
	item := newProjectListItem(meta, false, "alpha")
	if got := item.Tag(); got != "alpha" {
		t.Fatalf("expected tag alpha, got %s", got)
	}
	filter := item.FilterValue()
	if !strings.Contains(filter, "alpha") {
		t.Fatalf("expected filter value %q to include tag", filter)
	}
	// Ensure metadata tags were copied
	meta.Tags[0] = "mutated"
	if strings.Contains(strings.Join(item.meta.Tags, ","), "mutated") {
		t.Fatalf("project list item should copy metadata tags to avoid mutation")
	}
}

func TestActiveProjectStatusFormatting(t *testing.T) {
	cfg := &config.Config{TaskMasterPath: "/work/project"}
	model := Model{config: cfg}

	if status := model.activeProjectStatus(); status == "" {
		t.Fatal("expected fallback status when only config path is set")
	}

	model.activeProject = &projects.Metadata{Name: "Workspace"}
	model.config.ActiveTag = "alpha"
	status := model.activeProjectStatus()
	expected := "Active: Workspace [alpha]"
	if status != expected {
		t.Fatalf("expected %q, got %q", expected, status)
	}
}
