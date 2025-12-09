package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/adriangreen/tm-tui/internal/executor"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
	"github.com/adriangreen/tm-tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tm-tui",
		Short: "Task Master TUI - Interactive terminal interface for Task Master AI",
		Long: `Task Master TUI is a terminal user interface for managing and executing
development tasks with Task Master AI. It provides an interactive way to
view, organize, and track your project tasks.`,
		RunE: runTUI,
	}
	
	// Add flags
	cmd.PersistentFlags().Bool("clear-state", false, "Clear the TUI state before starting")

	return cmd
}

// runTUI starts the Bubble Tea TUI application
func runTUI(cmd *cobra.Command, args []string) error {
	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Create config manager
	configManager, err := config.NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	cfg := configManager.GetConfig()

	// Check if --clear-state flag is set
	clearState, _ := cmd.Flags().GetBool("clear-state")
	if clearState && cfg.StatePath != "" {
		if err := os.Remove(cfg.StatePath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to clear state file: %v\n", err)
		} else if err == nil {
			fmt.Fprintf(os.Stderr, "TUI state cleared successfully\n")
		}
	}

	// Check if taskmaster path is found
	if cfg.TaskMasterPath == "" {
		fmt.Fprintf(os.Stderr, "Error: .taskmaster directory not found\n\n")
		fmt.Fprintf(os.Stderr, "Task Master TUI requires a Task Master project.\n")
		fmt.Fprintf(os.Stderr, "Please initialize a Task Master project first:\n\n")
		fmt.Fprintf(os.Stderr, "  task-master init\n\n")
		return fmt.Errorf("no .taskmaster directory found")
	}

	// Initialize Task Master service
	taskService, err := taskmaster.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize task service: %w", err)
	}

	// Start file watchers
	if err := taskService.StartWatcher(ctx); err != nil {
		// Log warning but don't fail - watchers are optional
		fmt.Fprintf(os.Stderr, "Warning: failed to start task watcher: %v\n", err)
	}
	defer taskService.StopWatcher()

	if err := configManager.StartWatcher(ctx); err != nil {
		// Log warning but don't fail - watchers are optional
		fmt.Fprintf(os.Stderr, "Warning: failed to start config watcher: %v\n", err)
	}
	defer configManager.StopWatcher()

	// Initialize executor service
	execService, err := executor.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize executor service: %w", err)
	}
	defer execService.Close()

	// Create and run the TUI
	m := ui.NewModel(cfg, configManager, taskService, execService)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
