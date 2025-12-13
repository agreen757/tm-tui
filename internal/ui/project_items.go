package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/adriangreen/tm-tui/internal/projects"
)

type projectTagItem struct {
	summary projects.TagSummary
	active  bool
}

func newProjectTagItem(summary projects.TagSummary, active bool) *projectTagItem {
	return &projectTagItem{summary: summary, active: active}
}

func (i *projectTagItem) Title() string {
	prefix := "  "
	if i.active {
		prefix = "● "
	}
	return fmt.Sprintf("%s%s", prefix, i.summary.Name)
}

func (i *projectTagItem) Description() string {
	lastUsed := "never"
	if !i.summary.LastUsed.IsZero() {
		lastUsed = formatRelativeTime(i.summary.LastUsed)
	}
	return fmt.Sprintf("%d project(s) • last used %s", i.summary.ProjectCount, lastUsed)
}

func (i *projectTagItem) FilterValue() string {
	return i.summary.Name
}

type projectListItem struct {
	meta   *projects.Metadata
	active bool
	tag    string
}

func newProjectListItem(meta *projects.Metadata, active bool, tag string) *projectListItem {
	clone := *meta
	if len(meta.Tags) > 0 {
		clone.Tags = append([]string{}, meta.Tags...)
	}
	return &projectListItem{meta: &clone, active: active, tag: strings.TrimSpace(tag)}
}

func (i *projectListItem) Title() string {
	if i.meta == nil {
		return ""
	}
	prefix := "  "
	if i.active {
		prefix = "● "
	}
	display := i.meta.Name
	if display == "" {
		display = filepath.Base(i.meta.Path)
	}
	if display == "" {
		display = i.meta.Path
	}
	return fmt.Sprintf("%s%s", prefix, display)
}

func (i *projectListItem) Description() string {
	if i.meta == nil {
		return ""
	}
	parts := []string{}
	if len(i.meta.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("tags: %s", strings.Join(i.meta.Tags, ", ")))
	}
	if i.meta.Path != "" {
		parts = append(parts, i.meta.Path)
	}
	if !i.meta.LastUsed.IsZero() {
		parts = append(parts, fmt.Sprintf("last used %s", formatRelativeTime(i.meta.LastUsed)))
	}
	return strings.Join(parts, " • ")
}

func (i *projectListItem) FilterValue() string {
	if i.meta == nil {
		return ""
	}
	values := []string{i.meta.Name, i.meta.Path, strings.Join(i.meta.Tags, " ")}
	if i.tag != "" {
		values = append(values, i.tag)
	}
	return strings.ToLower(strings.Join(values, " "))
}

func (i *projectListItem) Tag() string {
	return i.tag
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	}
	if duration < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	}
	return t.Format("Jan 02, 15:04")
}
