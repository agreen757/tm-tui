package projects

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// defaultMRULimit caps how many recently used projects we surface.
	defaultMRULimit = 12
)

// Metadata captures information about a Task Master project directory.
type Metadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Tags        []string  `json:"tags"`
	LastUsed    time.Time `json:"lastUsed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Description string    `json:"description,omitempty"`
}

// PrimaryTag returns the first tag if present.
func (m *Metadata) PrimaryTag() string {
	if m == nil || len(m.Tags) == 0 {
		return ""
	}
	return m.Tags[0]
}

// HasTag checks whether the metadata contains a tag (case-insensitive).
func (m *Metadata) HasTag(tag string) bool {
	if m == nil || tag == "" {
		return false
	}
	needle := strings.ToLower(strings.TrimSpace(tag))
	for _, candidate := range m.Tags {
		if strings.ToLower(candidate) == needle {
			return true
		}
	}
	return false
}

// TagSummary aggregates tag usage statistics across projects.
type TagSummary struct {
	Name         string
	ProjectCount int
	LastUsed     time.Time
}

type registryFile struct {
	Projects []*Metadata `json:"projects"`
	MRU      []string    `json:"mru"`
}

// Registry persists metadata for discovered Task Master projects.
type Registry struct {
	path     string
	mu       sync.RWMutex
	projects map[string]*Metadata
	mru      []string
}

// Load reads a registry from disk (creating an empty one if the file is missing).
func Load(path string) (*Registry, error) {
	if path == "" {
		return nil, errors.New("registry path is empty")
	}

	reg := &Registry{
		path:     path,
		projects: make(map[string]*Metadata),
		mru:      []string{},
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := reg.ensureDir(); err != nil {
			return nil, err
		}
		return reg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read project registry: %w", err)
	}

	var file registryFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to parse project registry: %w", err)
	}

	for _, meta := range file.Projects {
		if meta == nil || meta.Path == "" {
			continue
		}
		meta.Path = normalizePath(meta.Path)
		reg.projects[meta.Path] = sanitizeMetadata(meta)
	}
	reg.mru = dedupePaths(file.MRU)
	return reg, nil
}

// Save writes the registry state to disk.
func (r *Registry) Save() error {
	if r == nil {
		return errors.New("registry is nil")
	}
	if r.path == "" {
		return errors.New("registry path is empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.ensureDir(); err != nil {
		return err
	}

	payload := registryFile{MRU: append([]string{}, r.mru...)}
	for _, meta := range r.projects {
		payload.Projects = append(payload.Projects, meta)
	}

	sort.Slice(payload.Projects, func(i, j int) bool {
		return payload.Projects[i].Name < payload.Projects[j].Name
	})

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode project registry: %w", err)
	}

	if err := os.WriteFile(r.path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write project registry: %w", err)
	}
	return nil
}

// Path returns the on-disk location backing this registry.
func (r *Registry) Path() string {
	if r == nil {
		return ""
	}
	return r.path
}

// RegisterProject ensures metadata exists for a given path and merges tags.
func (r *Registry) RegisterProject(meta *Metadata) *Metadata {
	if r == nil || meta == nil || meta.Path == "" {
		return nil
	}

	norm := normalizePath(meta.Path)
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.projects[norm]
	if !ok {
		cleaned := sanitizeMetadata(meta)
		if cleaned.ID == "" {
			cleaned.ID = hashPath(norm)
		}
		if cleaned.Name == "" {
			cleaned.Name = filepath.Base(norm)
		}
		if cleaned.CreatedAt.IsZero() {
			cleaned.CreatedAt = time.Now().UTC()
		}
		cleaned.UpdatedAt = time.Now().UTC()
		cleaned.Path = norm
		r.projects[norm] = cleaned
		return cleaned
	}

	mergedTags := mergeTags(existing.Tags, meta.Tags)
	existing.Tags = mergedTags
	existing.Name = pickNonEmpty(meta.Name, existing.Name)
	if desc := strings.TrimSpace(meta.Description); desc != "" {
		existing.Description = desc
	}
	existing.UpdatedAt = time.Now().UTC()
	return existing
}

// Projects returns every tracked project sorted by name.
func (r *Registry) Projects() []*Metadata {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Metadata, 0, len(r.projects))
	for _, meta := range r.projects {
		out = append(out, meta)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

// Get returns metadata for a project path.
func (r *Registry) Get(path string) (*Metadata, bool) {
	if r == nil || path == "" {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	meta, ok := r.projects[normalizePath(path)]
	return meta, ok
}

// Tags summarizes how many projects map to every tag name.
func (r *Registry) Tags() []TagSummary {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := make(map[string]*TagSummary)
	for _, meta := range r.projects {
		for _, tag := range meta.Tags {
			key := strings.ToLower(tag)
			summary, ok := counts[key]
			if !ok {
				counts[key] = &TagSummary{Name: tag, ProjectCount: 1, LastUsed: meta.LastUsed}
				continue
			}
			summary.ProjectCount++
			if meta.LastUsed.After(summary.LastUsed) {
				summary.LastUsed = meta.LastUsed
			}
		}
	}

	result := make([]TagSummary, 0, len(counts))
	for _, summary := range counts {
		result = append(result, *summary)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ProjectCount == result[j].ProjectCount {
			return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
		}
		return result[i].ProjectCount > result[j].ProjectCount
	})
	return result
}

// ProjectsForTag returns the projects associated with a tag.
func (r *Registry) ProjectsForTag(tag string) []*Metadata {
	if r == nil || tag == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*Metadata
	needle := strings.ToLower(strings.TrimSpace(tag))
	for _, meta := range r.projects {
		if meta.HasTag(needle) {
			filtered = append(filtered, meta)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		// Prefer most recently used first, fallback to name.
		if filtered[i].LastUsed.Equal(filtered[j].LastUsed) {
			return strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		}
		return filtered[i].LastUsed.After(filtered[j].LastUsed)
	})
	return filtered
}

// RecordUse updates MRU ordering and timestamps.
func (r *Registry) RecordUse(path string) *Metadata {
	if r == nil || path == "" {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	norm := normalizePath(path)
	meta, ok := r.projects[norm]
	if !ok {
		meta = &Metadata{Path: norm, Name: filepath.Base(norm), ID: hashPath(norm)}
		meta.CreatedAt = time.Now().UTC()
		r.projects[norm] = meta
	}
	now := time.Now().UTC()
	meta.LastUsed = now
	meta.UpdatedAt = now

	r.mru = prependAndDedupe(norm, r.mru, defaultMRULimit)
	return meta
}

// RecentProjects returns up to limit projects ordered by MRU.
func (r *Registry) RecentProjects(limit int) []*Metadata {
	if r == nil {
		return nil
	}
	if limit <= 0 {
		limit = defaultMRULimit
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*Metadata
	for _, path := range r.mru {
		if meta, ok := r.projects[path]; ok {
			results = append(results, meta)
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

// MergeTags ensures the given project contains the specified tags.
func (r *Registry) MergeTags(path string, tags []string) (*Metadata, bool) {
	if r == nil || path == "" || len(tags) == 0 {
		return nil, false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	meta, ok := r.projects[normalizePath(path)]
	if !ok {
		return nil, false
	}
	merged := mergeTags(meta.Tags, tags)
	if len(merged) == len(meta.Tags) {
		return meta, false
	}
	meta.Tags = merged
	meta.UpdatedAt = time.Now().UTC()
	return meta, true
}

// ensureDir guarantees the registry directory exists.
func (r *Registry) ensureDir() error {
	dir := filepath.Dir(r.path)
	if dir == "" || dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}
	return nil
}

func sanitizeMetadata(meta *Metadata) *Metadata {
	if meta == nil {
		return &Metadata{}
	}
	copy := *meta
	copy.Path = normalizePath(copy.Path)
	copy.Tags = sanitizeTags(copy.Tags)
	if copy.ID == "" && copy.Path != "" {
		copy.ID = hashPath(copy.Path)
	}
	return &copy
}

func sanitizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := make(map[string]string)
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; !ok {
			seen[key] = trimmed
		}
	}
	out := make([]string, 0, len(seen))
	for _, original := range seen {
		out = append(out, original)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i]) < strings.ToLower(out[j])
	})
	return out
}

func mergeTags(existing, incoming []string) []string {
	if len(existing) == 0 {
		return sanitizeTags(incoming)
	}
	combo := append([]string{}, existing...)
	combo = append(combo, incoming...)
	return sanitizeTags(combo)
}

func dedupePaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		norm := normalizePath(path)
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, norm)
	}
	return out
}

func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return abs
}

func hashPath(path string) string {
	sum := sha1.Sum([]byte(strings.ToLower(path)))
	return hex.EncodeToString(sum[:8])
}

func prependAndDedupe(value string, items []string, limit int) []string {
	filtered := []string{value}
	for _, existing := range items {
		if normalizePath(existing) == normalizePath(value) {
			continue
		}
		filtered = append(filtered, existing)
		if limit > 0 && len(filtered) >= limit {
			break
		}
	}
	return filtered
}

func pickNonEmpty(values ...string) string {
	for _, val := range values {
		if strings.TrimSpace(val) != "" {
			return val
		}
	}
	return ""
}
