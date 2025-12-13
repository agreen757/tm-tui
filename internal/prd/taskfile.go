package prd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type fileFormat int

const (
	formatUnknown fileFormat = iota
	formatTagged
	formatRoot
	formatArray
)

// TaskFileManager reads and writes Task Master tasks.json documents while preserving
// existing tags and metadata blocks.
type TaskFileManager struct {
	path      string
	activeTag string
}

// NewTaskFileManager creates a manager rooted at the provided Task Master directory.
func NewTaskFileManager(rootDir, activeTag string) *TaskFileManager {
	return &TaskFileManager{
		path:      filepath.Join(rootDir, ".taskmaster", "tasks", "tasks.json"),
		activeTag: activeTag,
	}
}

// Append adds tasks to the end of the current tasks array.
func (m *TaskFileManager) Append(tasks []map[string]interface{}) error {
	doc, err := m.load()
	if err != nil {
		return err
	}
	doc.append(tasks)
	return doc.save()
}

// Replace overwrites the tasks array with the provided tasks.
func (m *TaskFileManager) Replace(tasks []map[string]interface{}) error {
	doc, err := m.load()
	if err != nil {
		return err
	}
	doc.replace(tasks)
	return doc.save()
}

func (m *TaskFileManager) load() (*taskFileDocument, error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse tasks file: %w", err)
	}

	doc := &taskFileDocument{path: m.path}

	switch typed := raw.(type) {
	case map[string]interface{}:
		// Tagged or root map formats
		if tasks, ok := asInterfaceSlice(typed["tasks"]); ok {
			doc.format = formatRoot
			doc.root = typed
			doc.tasks = tasks
			return doc, nil
		}

		// Try preferred tag
		if entry, ok := doc.extractTagEntry(typed, m.activeTag); ok {
			doc.format = formatTagged
			doc.root = typed
			doc.tagKey = m.activeTag
			doc.tagEntry = entry
			doc.tasks = entry["tasks"].([]interface{})
			return doc, nil
		}

		// Look for any tag with tasks
		for key := range typed {
			entry, ok := doc.extractTagEntry(typed, key)
			if ok {
				doc.format = formatTagged
				doc.root = typed
				doc.tagKey = key
				doc.tagEntry = entry
				doc.tasks = entry["tasks"].([]interface{})
				return doc, nil
			}
		}
	case []interface{}:
		doc.format = formatArray
		doc.tasks = typed
		return doc, nil
	}

	return nil, fmt.Errorf("unrecognized tasks.json structure")
}

// taskFileDocument wraps different layout variations for tasks.json.
type taskFileDocument struct {
	format   fileFormat
	path     string
	root     map[string]interface{}
	tagEntry map[string]interface{}
	tagKey   string
	tasks    []interface{}
}

func (d *taskFileDocument) append(tasks []map[string]interface{}) {
	for _, t := range tasks {
		d.tasks = append(d.tasks, t)
	}
}

func (d *taskFileDocument) replace(tasks []map[string]interface{}) {
	d.tasks = make([]interface{}, len(tasks))
	for i, t := range tasks {
		d.tasks[i] = t
	}
}

func (d *taskFileDocument) save() error {
	switch d.format {
	case formatTagged:
		if d.tagEntry == nil {
			d.tagEntry = map[string]interface{}{}
		}
		d.tagEntry["tasks"] = d.tasks
		if d.root == nil {
			d.root = map[string]interface{}{}
		}
		d.root[d.tagKey] = d.tagEntry
		data, err := json.MarshalIndent(d.root, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize tasks file: %w", err)
		}
		return os.WriteFile(d.path, append(data, '\n'), 0644)
	case formatRoot:
		if d.root == nil {
			d.root = map[string]interface{}{}
		}
		d.root["tasks"] = d.tasks
		data, err := json.MarshalIndent(d.root, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize tasks file: %w", err)
		}
		return os.WriteFile(d.path, append(data, '\n'), 0644)
	case formatArray:
		data, err := json.MarshalIndent(d.tasks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize tasks file: %w", err)
		}
		return os.WriteFile(d.path, append(data, '\n'), 0644)
	default:
		return fmt.Errorf("unsupported tasks file format")
	}
}

func (d *taskFileDocument) extractTagEntry(root map[string]interface{}, key string) (map[string]interface{}, bool) {
	if key == "" {
		return nil, false
	}
	value, ok := root[key]
	if !ok {
		return nil, false
	}
	entry, ok := value.(map[string]interface{})
	if !ok {
		return nil, false
	}
	tasks, ok := asInterfaceSlice(entry["tasks"])
	if !ok {
		return nil, false
	}
	entry["tasks"] = tasks
	return entry, true
}

func asInterfaceSlice(value interface{}) ([]interface{}, bool) {
	if value == nil {
		return nil, false
	}
	slice, ok := value.([]interface{})
	return slice, ok
}
