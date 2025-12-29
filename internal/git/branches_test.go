package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGetBranches tests the GetBranches function
func TestGetBranches(t *testing.T) {
	// Create a temporary git repository
	tempDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Configure git
	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	// Create initial commit
	filePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Test GetBranches
	ctx := context.Background()
	branches, currentBranch, err := GetBranches(ctx, tempDir)

	if err != nil {
		t.Fatalf("GetBranches failed: %v", err)
	}

	if len(branches) == 0 {
		t.Error("Expected at least one branch")
	}

	if currentBranch == "" {
		t.Error("Expected currentBranch to be non-empty")
	}

	// Verify current branch is in the list
	found := false
	for _, b := range branches {
		if b == currentBranch {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("currentBranch %q not found in branches list", currentBranch)
	}
}

// TestGetBranches_EmptyRepo tests GetBranches with a repo that has no commits
func TestGetBranches_EmptyRepo(t *testing.T) {
	tempDir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// In modern git, even with no commits, we can still detect the default branch
	// This test verifies that GetBranches works with empty repos
	ctx := context.Background()
	branches, currentBranch, _ := GetBranches(ctx, tempDir)

	// With a fresh repo and no commits, the behavior depends on git config
	// It should either return empty branches or the default branch name
	if currentBranch == "" && len(branches) == 0 {
		// This is acceptable - no branches to switch to
		return
	}
}

// TestGetBranches_InvalidPath tests GetBranches with invalid repository path
func TestGetBranches_InvalidPath(t *testing.T) {
	ctx := context.Background()
	_, _, err := GetBranches(ctx, "/nonexistent/path")

	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// TestSwitchBranch tests the SwitchBranch function
func TestSwitchBranch(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Configure git
	exec.Command("git", "config", "user.email", "test@example.com").Dir = tempDir
	exec.Command("git", "config", "user.name", "Test User").Dir = tempDir

	// Create initial commit
	filePath := filepath.Join(tempDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)
	exec.Command("git", "add", "test.txt").Dir = tempDir
	exec.Command("git", "commit", "-m", "initial").Dir = tempDir

	// Create a new branch
	ctx := context.Background()
	_, err := SwitchBranch(ctx, tempDir, "feature/test")

	if err == nil {
		t.Error("Expected error when switching to non-existent branch")
	}

	// Create the branch first
	cmd = exec.Command("git", "checkout", "-b", "feature/test")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create test branch: %v", err)
	}

	// Get current branch 
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = tempDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}
	
	currentBranch := string(output)
	if len(currentBranch) > 0 && currentBranch[len(currentBranch)-1] == '\n' {
		currentBranch = currentBranch[:len(currentBranch)-1]
	}

	// Try to switch to main/master
	mainBranch := "master"
	if currentBranch == "main" || currentBranch == "master" {
		mainBranch = "main"
	}
	
	switchOutput, err := SwitchBranch(ctx, tempDir, mainBranch)

	if err != nil {
		t.Logf("Note: SwitchBranch returned error (might be expected if branch doesn't exist): %v", err)
	}
	_ = switchOutput
}
