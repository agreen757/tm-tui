package taskmaster

import "testing"

func TestExpandTaskDraftsFallback(t *testing.T) {
	task := &Task{ID: "1", Title: "Implement auth", Description: ""}
	drafts := ExpandTaskDrafts(task, ExpandTaskOptions{Depth: 1})
	if len(drafts) == 0 {
		t.Fatalf("expected fallback draft")
	}
}
