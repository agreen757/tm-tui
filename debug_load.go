package main

import (
	"fmt"
	"os"

	"github.com/adriangreen/tm-tui/internal/config"
	"github.com/adriangreen/tm-tui/internal/taskmaster"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config TaskMasterPath: '%s'\n", cfg.TaskMasterPath)
	
	// Test detection directly
	root, err := taskmaster.FindTaskmasterRoot()
	fmt.Printf("Direct FindTaskmasterRoot: root='%s' err=%v\n", root, err)

	svc, err := taskmaster.NewService(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Service error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Available: %v\n", svc.IsAvailable())
	fmt.Printf("Root Dir: %s\n", svc.RootDir)

	tasks, warnings := svc.GetTasks()
	fmt.Printf("Number of tasks: %d\n", len(tasks))
	fmt.Printf("Warnings: %v\n", warnings)

	for i, task := range tasks {
		fmt.Printf("Task %d: ID='%s' Title='%s' Status='%s'\n", i, task.ID, task.Title, task.Status)
	}

	counts := svc.GetTaskCount()
	fmt.Printf("Task counts: %+v\n", counts)
}
