package projects

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// AutoDiscover scans the provided paths (and their immediate children) for
// Task Master projects, registering any newly found entries.
func (r *Registry) AutoDiscover(paths []string) (int, error) {
	if r == nil {
		return 0, nil
	}

	unique := uniquePaths(paths)
	seen := make(map[string]struct{})
	added := 0

	for _, root := range unique {
		candidates := discoverProjectDirs(root)
		for _, candidate := range candidates {
			norm := normalizePath(candidate)
			if norm == "" {
				continue
			}
			if _, already := seen[norm]; already {
				continue
			}
			seen[norm] = struct{}{}

			if _, exists := r.Get(norm); exists {
				// Update tags in case they changed on disk.
				tags := InferTags(norm)
				r.MergeTags(norm, tags)
				continue
			}

			meta := &Metadata{
				Path: norm,
				Name: filepath.Base(norm),
				Tags: InferTags(norm),
			}
			r.RegisterProject(meta)
			added++
		}
	}

	if added > 0 {
		if err := r.Save(); err != nil {
			return added, err
		}
	}

	return added, nil
}

// InferTags derives reasonable tags for a project directory by inspecting
// .taskmaster metadata combined with folder context and optional extras.
func InferTags(projectPath string, extras ...string) []string {
	var tags []string
	if tag := readCurrentTag(projectPath); tag != "" {
		tags = append(tags, tag)
	}
	tags = append(tags, extras...)
	tags = append(tags, filepath.Base(projectPath))

	parent := filepath.Base(filepath.Dir(projectPath))
	if parent != "" && parent != "." {
		tags = append(tags, parent)
	}

	return sanitizeTags(tags)
}

func discoverProjectDirs(root string) []string {
	if root == "" {
		return nil
	}

	var results []string
	if isTaskmasterProject(root) {
		results = append(results, root)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return results
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(root, entry.Name())
		if isTaskmasterProject(candidate) {
			results = append(results, candidate)
		}
	}

	return results
}

func isTaskmasterProject(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(path, ".taskmaster"))
	if err != nil || !info.IsDir() {
		return false
	}
	_, err = os.Stat(filepath.Join(path, ".taskmaster", "tasks", "tasks.json"))
	return err == nil
}

func readCurrentTag(path string) string {
	statePath := filepath.Join(path, ".taskmaster", "state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return ""
	}
	var payload struct {
		CurrentTag string `json:"currentTag"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.CurrentTag)
}

func uniquePaths(paths []string) []string {
	set := make(map[string]struct{})
	for _, path := range paths {
		norm := normalizePath(path)
		if norm == "" {
			continue
		}
		set[norm] = struct{}{}
		// Also include immediate parent to widen discovery scope slightly.
		parent := filepath.Dir(norm)
		if parent != "" && parent != norm {
			set[parent] = struct{}{}
		}
	}

	out := make([]string, 0, len(set))
	for path := range set {
		out = append(out, path)
	}
	return out
}

// WalkProjects executes fn for each project located beneath the supplied roots
// (only visiting the root and its first-level children).
func WalkProjects(roots []string, fn func(path string) error) error {
	if fn == nil {
		return nil
	}
	for _, root := range uniquePaths(roots) {
		paths := discoverProjectDirs(root)
		for _, candidate := range paths {
			if err := fn(candidate); err != nil {
				if errors.Is(err, fs.SkipDir) {
					continue
				}
				return err
			}
		}
	}
	return nil
}
