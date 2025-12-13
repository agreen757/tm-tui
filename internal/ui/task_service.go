package ui

import (
	"context"

	"github.com/adriangreen/tm-tui/internal/projects"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
)

// TaskService defines the subset of taskmaster.Service behavior required by the UI.
type TaskService interface {
	GetTasks() ([]taskmaster.Task, []string)
	LoadTasks(ctx context.Context) error
	ReloadEvents() <-chan struct{}
	AnalyzeComplexity(ctx context.Context, scope string, taskID string, tags []string) (*taskmaster.ComplexityReport, error)
	AnalyzeComplexityWithProgress(ctx context.Context, scope string, taskID string, tags []string, onProgress func(taskmaster.ComplexityProgressState)) (*taskmaster.ComplexityReport, error)
	ParsePRDWithProgress(ctx context.Context, inputPath string, mode taskmaster.ParsePrdMode, onProgress func(taskmaster.ParsePrdProgressState)) error
	ExpandTaskWithProgress(ctx context.Context, taskID string, opts taskmaster.ExpandTaskOptions, prompt string, onProgress func(taskmaster.ExpandProgressState)) error
	GetLatestComplexityReport() *taskmaster.ComplexityReport
	ExportComplexityReport(ctx context.Context, format string, outputPath string) (string, error)
	IsAvailable() bool
	AnalyzeDeleteImpact(taskIDs []string, opts taskmaster.DeleteOptions) (*taskmaster.DeleteImpact, error)
	DeleteTasks(ctx context.Context, taskIDs []string, opts taskmaster.DeleteOptions) (*taskmaster.DeleteResult, error)
	UndoAction(ctx context.Context, actionID string) error
	ListTagContexts(ctx context.Context, includeMetadata bool) (*taskmaster.TagList, error)
	AddTagContext(ctx context.Context, opts taskmaster.TagAddOptions) (*taskmaster.TagOperationResult, error)
	DeleteTagContext(ctx context.Context, name string, skipConfirmation bool) (*taskmaster.TagOperationResult, error)
	UseTagContext(ctx context.Context, name string) (*taskmaster.TagOperationResult, error)
	RenameTagContext(ctx context.Context, oldName, newName string) (*taskmaster.TagOperationResult, error)
	CopyTagContext(ctx context.Context, sourceName, targetName string, opts taskmaster.TagCopyOptions) (*taskmaster.TagOperationResult, error)
	ProjectRegistry() *projects.Registry
	ActiveProjectMetadata() *projects.Metadata
	SwitchProject(ctx context.Context, projectPath string) (*projects.Metadata, error)
	DiscoverProjects(ctx context.Context, roots []string) (int, error)
}
