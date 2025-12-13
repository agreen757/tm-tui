package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/adriangreen/tm-tui/internal/projects"
	"github.com/adriangreen/tm-tui/internal/ui/dialog"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openProjectTagsDialog() tea.Cmd {
	registry := m.taskService.ProjectRegistry()
	if registry == nil {
		m.showErrorDialog("Project Tags", "Project registry not available.")
		return nil
	}

	m.refreshProjectRegistry()

	summaries := registry.Tags()
	if len(summaries) == 0 {
		m.showErrorDialog("Project Tags", "No tags discovered yet. Add tags via Task Master to organize projects.")
		return nil
	}

	activeTag := ""
	if m.activeProject != nil {
		activeTag = m.activeProject.PrimaryTag()
	}

	items := make([]dialog.ListItem, 0, len(summaries))
	for _, summary := range summaries {
		items = append(items, newProjectTagItem(summary, activeTag != "" && summary.Name == activeTag))
	}

	width := clampInt(m.width-12, 60, 100)
	height := clampInt(m.height-8, 18, 30)
	dlg := dialog.NewListDialog("Project Tags", width, height, items)
	dlg.SetShowDescription(true)
	dlg.EnableFiltering("Type to filter tags...")
	m.appState.AddDialog(dlg, nil)
	return nil
}

func (m *Model) openProjectSelectionDialog(tag string) tea.Cmd {
	registry := m.taskService.ProjectRegistry()
	if registry == nil {
		return nil
	}
	projects := registry.ProjectsForTag(tag)
	if len(projects) == 0 {
		m.showErrorDialog("Projects", fmt.Sprintf("No projects found for tag '%s'", tag))
		return nil
	}

	items := make([]dialog.ListItem, 0, len(projects))
	currentPath := ""
	if m.activeProject != nil {
		currentPath = m.activeProject.Path
	}
	for _, meta := range projects {
		items = append(items, newProjectListItem(meta, pathsEqual(meta.Path, currentPath), tag))
	}

	width := clampInt(m.width-12, 70, 110)
	height := clampInt(m.height-6, 20, 32)
	dlg := dialog.NewListDialog(fmt.Sprintf("Projects tagged '%s'", tag), width, height, items)
	dlg.SetShowDescription(true)
	dlg.EnableFiltering("Search projects...")
	m.appState.AddDialog(dlg, nil)
	return nil
}

func (m *Model) openQuickProjectSwitchDialog() tea.Cmd {
	m.refreshProjectRegistry()
	registry := m.taskService.ProjectRegistry()
	if registry == nil {
		m.showErrorDialog("Quick Switch", "Project registry not available.")
		return nil
	}

	recent := registry.RecentProjects(12)
	if len(recent) == 0 {
		m.showErrorDialog("Quick Switch", "No recent projects recorded yet.")
		return nil
	}

	items := make([]dialog.ListItem, 0, len(recent))
	current := ""
	if m.activeProject != nil {
		current = m.activeProject.Path
	}
	for _, meta := range recent {
		items = append(items, newProjectListItem(meta, pathsEqual(meta.Path, current), meta.PrimaryTag()))
	}

	width := clampInt(m.width-8, 70, 110)
	height := clampInt(m.height-8, 18, 28)
	dlg := dialog.NewListDialog("Quick Project Switch", width, height, items)
	dlg.SetShowDescription(true)
	dlg.EnableFiltering("Filter recent projects...")
	m.appState.AddDialog(dlg, nil)
	return nil
}

func (m *Model) openProjectSearchDialog() tea.Cmd {
	m.refreshProjectRegistry()
	registry := m.taskService.ProjectRegistry()
	if registry == nil {
		m.showErrorDialog("Project Search", "Project registry not available.")
		return nil
	}

	m.refreshProjectRegistry()

	tagItems := registry.Tags()
	projectItems := registry.Projects()

	if len(tagItems) == 0 && len(projectItems) == 0 {
		m.showErrorDialog("Project Search", "No project metadata discovered yet.")
		return nil
	}

	items := make([]dialog.ListItem, 0, len(tagItems)+len(projectItems))
	activeTag := ""
	activePath := ""
	if m.activeProject != nil {
		activeTag = m.activeProject.PrimaryTag()
		activePath = m.activeProject.Path
	}

	for _, summary := range tagItems {
		items = append(items, newProjectTagItem(summary, activeTag != "" && summary.Name == activeTag))
	}
	for _, meta := range projectItems {
		items = append(items, newProjectListItem(meta, pathsEqual(meta.Path, activePath), meta.PrimaryTag()))
	}

	width := clampInt(m.width-8, 70, 110)
	height := clampInt(m.height-4, 24, 36)
	dlg := dialog.NewListDialog("Search Projects & Tags", width, height, items)
	dlg.SetShowDescription(true)
	dlg.EnableFiltering("Search tags or projects...")
	m.appState.AddDialog(dlg, nil)
	return nil
}

func (m *Model) refreshProjectRegistry() {
	roots := m.projectDiscoveryRoots()
	if len(roots) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := m.taskService.DiscoverProjects(ctx, roots); err != nil {
		m.addLogLine(fmt.Sprintf("Project discovery error: %v", err))
	}
	if registry := m.taskService.ProjectRegistry(); registry != nil {
		m.projectRegistry = registry
	}
}

func (m *Model) requestProjectSwitch(meta *projects.Metadata, preferredTag string) tea.Cmd {
	if meta == nil || meta.Path == "" {
		return nil
	}
	preferredTag = strings.TrimSpace(preferredTag)
	if preferredTag == "" {
		preferredTag = meta.PrimaryTag()
	}
	if m.activeProject != nil && pathsEqual(m.activeProject.Path, meta.Path) {
		if preferredTag != "" {
			return m.useProjectTag(preferredTag)
		}
		m.addLogLine(fmt.Sprintf("Already on project %s", meta.Name))
		return nil
	}
	if m.execService != nil && m.execService.IsRunning() {
		m.showErrorDialog("Switch Project", "Wait for the current Task Master command to finish before switching projects.")
		return nil
	}

	if m.hasPendingState() {
		m.pendingProjectSwitch = meta
		m.pendingProjectTag = preferredTag
		message := fmt.Sprintf("Switch to project '%s'? Current selections and filters will be cleared.", meta.Name)
		confirm := dialog.YesNo("Switch Project", message, false)
		m.appState.AddDialog(confirm, func(_ interface{}, _ error) tea.Cmd {
			if confirm.Result() == dialog.ConfirmationResultYes {
				return m.executeProjectSwitch(meta, preferredTag)
			}
			m.pendingProjectSwitch = nil
			m.pendingProjectTag = ""
			return nil
		})
		return nil
	}

	return m.executeProjectSwitch(meta, preferredTag)
}

func (m *Model) executeProjectSwitch(meta *projects.Metadata, preferredTag string) tea.Cmd {
	if meta == nil || meta.Path == "" {
		return nil
	}
	targetPath := meta.Path
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		updated, err := m.taskService.SwitchProject(ctx, targetPath)
		if err != nil {
			return projectSwitchedMsg{Err: err, Tag: strings.TrimSpace(preferredTag)}
		}
		return projectSwitchedMsg{Meta: updated, Tag: strings.TrimSpace(preferredTag)}
	}
}

func (m *Model) hasPendingState() bool {
	if len(m.selectedIDs) > 0 || m.statusFilter != "" || m.searchQuery != "" || m.commandMode || m.searchMode {
		return true
	}
	return false
}

func (m *Model) projectDiscoveryRoots() []string {
	roots := []string{}
	if m.config != nil && m.config.TaskMasterPath != "" {
		roots = append(roots, m.config.TaskMasterPath)
		parent := filepath.Dir(m.config.TaskMasterPath)
		if parent != "" && parent != m.config.TaskMasterPath {
			roots = append(roots, parent)
		}
	}
	if m.activeProject != nil && m.activeProject.Path != "" {
		roots = append(roots, m.activeProject.Path)
	}
	return dedupeStrings(roots)
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func pathsEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
