package pathutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agreen757/tm-tui/internal/config"
)

// TestResolvePrdDirectoryPath_WithLastUsedPath tests that lastUsedPath takes priority
func TestResolvePrdDirectoryPath_WithLastUsedPath(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.Config
		lastUsedPath string
		expected     string
	}{
		{
			name:         "lastUsedPath with no config",
			cfg:          nil,
			lastUsedPath: "/some/custom/path",
			expected:     "/some/custom/path",
		},
		{
			name: "lastUsedPath overrides config",
			cfg: &config.Config{
				TaskMasterPath: "/taskmaster",
			},
			lastUsedPath: "/custom/prd/path",
			expected:     "/custom/prd/path",
		},
		{
			name:         "relative lastUsedPath",
			cfg:          nil,
			lastUsedPath: "relative/prd/path",
			expected:     "relative/prd/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolvePrdDirectoryPath(tt.cfg, tt.lastUsedPath)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestResolvePrdDirectoryPath_WithNilConfig tests behavior with nil config
func TestResolvePrdDirectoryPath_WithNilConfig(t *testing.T) {
	result := ResolvePrdDirectoryPath(nil, "")
	expected := filepath.Join(".taskmaster", "docs")
	if result != expected {
		t.Errorf("with nil config expected %q, got %q", expected, result)
	}
}

// TestResolvePrdDirectoryPath_WithEmptyTaskMasterPath tests behavior with empty TaskMasterPath
func TestResolvePrdDirectoryPath_WithEmptyTaskMasterPath(t *testing.T) {
	cfg := &config.Config{
		TaskMasterPath: "",
	}
	result := ResolvePrdDirectoryPath(cfg, "")
	expected := filepath.Join(".taskmaster", "docs")
	if result != expected {
		t.Errorf("with empty TaskMasterPath expected %q, got %q", expected, result)
	}
}

// TestResolvePrdDirectoryPath_DefaultLocation tests default fallback
func TestResolvePrdDirectoryPath_DefaultLocation(t *testing.T) {
	cfg := &config.Config{
		TaskMasterPath: "/nonexistent/path",
	}
	result := ResolvePrdDirectoryPath(cfg, "")
	expected := filepath.Join(".taskmaster", "docs")
	if result != expected {
		t.Errorf("with nonexistent TaskMasterPath expected default %q, got %q", expected, result)
	}
}

// TestResolvePrdDirectoryPath_ExistingDocsDirectory tests that existing .taskmaster/docs is used
func TestResolvePrdDirectoryPath_ExistingDocsDirectory(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, ".taskmaster", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	cfg := &config.Config{
		TaskMasterPath: tempDir,
	}
	result := ResolvePrdDirectoryPath(cfg, "")
	expected := docsDir

	if result != expected {
		t.Errorf("with existing docs directory expected %q, got %q", expected, result)
	}
}

// TestResolvePrdDirectoryPath_ExistingTaskMasterPath tests fallback to TaskMasterPath when docs doesn't exist
func TestResolvePrdDirectoryPath_ExistingTaskMasterPath(t *testing.T) {
	// Create temporary directory (but not .taskmaster/docs)
	tempDir := t.TempDir()

	cfg := &config.Config{
		TaskMasterPath: tempDir,
	}
	result := ResolvePrdDirectoryPath(cfg, "")
	expected := tempDir

	if result != expected {
		t.Errorf("with existing TaskMasterPath (no docs) expected %q, got %q", expected, result)
	}
}

// TestResolvePrdDirectoryPath_PriorityOrder tests the correct priority order
func TestResolvePrdDirectoryPath_PriorityOrder(t *testing.T) {
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, ".taskmaster", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	tests := []struct {
		name         string
		cfg          *config.Config
		lastUsedPath string
		expected     string
		description  string
	}{
		{
			name: "lastUsedPath has priority over existing .taskmaster/docs",
			cfg: &config.Config{
				TaskMasterPath: tempDir,
			},
			lastUsedPath: "/override/path",
			expected:     "/override/path",
			description:  "lastUsedPath should take priority",
		},
		{
			name: "existing .taskmaster/docs has priority over TaskMasterPath",
			cfg: &config.Config{
				TaskMasterPath: tempDir,
			},
			lastUsedPath: "",
			expected:     docsDir,
			description:  ".taskmaster/docs should be used if it exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolvePrdDirectoryPath(tt.cfg, tt.lastUsedPath)
			if result != tt.expected {
				t.Errorf("%s: expected %q, got %q", tt.description, tt.expected, result)
			}
		})
	}
}

// TestGetPrdDirectory_CreatesMissingDirectory tests that GetPrdDirectory creates directories
func TestGetPrdDirectory_CreatesMissingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	expectedPath := filepath.Join(tempDir, "new", "prd", "dir")

	cfg := &config.Config{
		TaskMasterPath: filepath.Join(tempDir, "nonexistent"),
	}

	// Should create the default .taskmaster/docs path
	path, err := GetPrdDirectory(cfg, expectedPath)
	if err != nil {
		t.Fatalf("GetPrdDirectory failed: %v", err)
	}

	if path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, path)
	}

	// Verify directory was created
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("created path is not a directory")
	}
}

// TestGetPrdDirectory_WithExistingDirectory tests that GetPrdDirectory works with existing directory
func TestGetPrdDirectory_WithExistingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	// Create .taskmaster/docs so it exists
	docsDir := filepath.Join(tempDir, ".taskmaster", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	cfg := &config.Config{
		TaskMasterPath: tempDir,
	}

	path, err := GetPrdDirectory(cfg, "")
	if err != nil {
		t.Fatalf("GetPrdDirectory failed: %v", err)
	}

	// Should return .taskmaster/docs which exists
	expectedPath := filepath.Join(tempDir, ".taskmaster", "docs")
	if path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, path)
	}

	// Verify directory exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory does not exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("path is not a directory")
	}
}

// TestGetPrdDirectory_CreatesWithDefaultPath tests directory creation with default path
func TestGetPrdDirectory_CreatesWithDefaultPath(t *testing.T) {
	// Change to temporary directory for this test
	originalCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(originalCwd)

	// Nil config should use default location
	path, err := GetPrdDirectory(nil, "")
	if err != nil {
		t.Fatalf("GetPrdDirectory with nil config failed: %v", err)
	}

	expectedPath := filepath.Join(".taskmaster", "docs")
	// Convert to absolute path for comparison since GetPrdDirectory returns relative path
	absPath, _ := filepath.Abs(path)
	expectedAbsPath, _ := filepath.Abs(expectedPath)

	if absPath != expectedAbsPath {
		t.Errorf("expected path %q, got %q", expectedAbsPath, absPath)
	}

	// Verify directory was created
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("created path is not a directory")
	}
}

// TestResolvePrdDirectoryPath_EmptyLastUsedPath treats empty string as not provided
func TestResolvePrdDirectoryPath_EmptyLastUsedPath(t *testing.T) {
	cfg := &config.Config{
		TaskMasterPath: "",
	}
	result := ResolvePrdDirectoryPath(cfg, "")
	expected := filepath.Join(".taskmaster", "docs")
	if result != expected {
		t.Errorf("empty lastUsedPath should be ignored, expected %q, got %q", expected, result)
	}
}

// BenchmarkResolvePrdDirectoryPath benchmarks the function
func BenchmarkResolvePrdDirectoryPath(b *testing.B) {
	cfg := &config.Config{
		TaskMasterPath: "/some/path",
	}
	for i := 0; i < b.N; i++ {
		ResolvePrdDirectoryPath(cfg, "")
	}
}

// BenchmarkGetPrdDirectory benchmarks the function (with temporary directory)
func BenchmarkGetPrdDirectory(b *testing.B) {
	tempDir := b.TempDir()
	cfg := &config.Config{
		TaskMasterPath: tempDir,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPrdDirectory(cfg, "")
	}
}
