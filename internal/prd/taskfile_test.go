package prd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTaskFileManagerAppendTagged(t *testing.T) {
	dir := t.TempDir()
	tasksDir := filepath.Join(dir, ".taskmaster", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	seed := map[string]interface{}{
		"tm-test": map[string]interface{}{
			"tasks": []interface{}{
				map[string]interface{}{"id": 1, "title": "Existing", "status": "pending"},
			},
			"metadata": map[string]interface{}{"created": "now"},
		},
	}
	data, _ := json.Marshal(seed)
	if err := os.WriteFile(filepath.Join(tasksDir, "tasks.json"), data, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	mgr := NewTaskFileManager(dir, "tm-test")
	doc := []map[string]interface{}{{"id": 2, "title": "New", "status": "pending"}}
	if err := mgr.Append(doc); err != nil {
		t.Fatalf("append: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(tasksDir, "tasks.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	entry := parsed["tm-test"].(map[string]interface{})
	tasks := entry["tasks"].([]interface{})
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestTaskFileManagerReplaceRoot(t *testing.T) {
	dir := t.TempDir()
	tasksDir := filepath.Join(dir, ".taskmaster", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	seed := map[string]interface{}{
		"tasks": []interface{}{
			map[string]interface{}{"id": 1, "title": "Existing"},
		},
	}
	data, _ := json.Marshal(seed)
	if err := os.WriteFile(filepath.Join(tasksDir, "tasks.json"), data, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	mgr := NewTaskFileManager(dir, "")
	if err := mgr.Replace([]map[string]interface{}{{"id": 99, "title": "Only"}}); err != nil {
		t.Fatalf("replace: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(tasksDir, "tasks.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	tasks := parsed["tasks"].([]interface{})
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	id := int(tasks[0].(map[string]interface{})["id"].(float64))
	if id != 99 {
		t.Fatalf("expected id 99, got %d", id)
	}
}
