package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGetCrushConfigPathEnvironmentOverride tests CRUSH_CONFIG_PATH env var
func TestGetCrushConfigPathEnvironmentOverride(t *testing.T) {
	tmpFile := t.TempDir() + "/custom.json"

	origVal := os.Getenv(EnvCrushConfigPath)
	defer os.Setenv(EnvCrushConfigPath, origVal)

	os.Setenv(EnvCrushConfigPath, tmpFile)

	path, err := GetCrushConfigPath()
	if err != nil {
		t.Fatalf("GetCrushConfigPath failed: %v", err)
	}

	if path != tmpFile {
		t.Errorf("Expected %q, got %q", tmpFile, path)
	}
}

// TestGetCrushConfigPathProjectRootOverride tests CRUSH_PROJECT_ROOT env var
func TestGetCrushConfigPathProjectRootOverride(t *testing.T) {
	tmpDir := t.TempDir()

	origVal := os.Getenv(EnvCrushProjectRoot)
	defer os.Setenv(EnvCrushProjectRoot, origVal)

	os.Setenv(EnvCrushProjectRoot, tmpDir)

	path, err := GetCrushConfigPath()
	if err != nil {
		t.Fatalf("GetCrushConfigPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".crush.json")
	if path != expected {
		t.Errorf("Expected %q, got %q", expected, path)
	}
}

// TestDetectProjectRootWithGit tests project root detection with .git
func TestDetectProjectRootWithGit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git: %v", err)
	}

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Change to subdirectory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(subDir)

	root, err := DetectProjectRoot()
	if err != nil {
		t.Fatalf("DetectProjectRoot failed: %v", err)
	}

	if !pathsEqual(root, tmpDir) {
		t.Errorf("Expected root %q, got %q", tmpDir, root)
	}
}

// TestDetectProjectRootWithGoMod tests project root detection with go.mod
func TestDetectProjectRootWithGoMod(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go.mod file
	modFile := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modFile, []byte("module test\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tmpDir, "internal", "config")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(nestedDir)

	root, err := DetectProjectRoot()
	if err != nil {
		t.Fatalf("DetectProjectRoot failed: %v", err)
	}

	if !pathsEqual(root, tmpDir) {
		t.Errorf("Expected root %q, got %q", tmpDir, root)
	}
}

// TestDetectProjectRootWithTaskmaster tests detection with .taskmaster
func TestDetectProjectRootWithTaskmaster(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .taskmaster directory
	tmDir := filepath.Join(tmpDir, ".taskmaster")
	if err := os.MkdirAll(tmDir, 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "tasks")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create tasks: %v", err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(subDir)

	root, err := DetectProjectRoot()
	if err != nil {
		t.Fatalf("DetectProjectRoot failed: %v", err)
	}

	if !pathsEqual(root, tmpDir) {
		t.Errorf("Expected root %q, got %q", tmpDir, root)
	}
}

// TestDetectProjectRootNoMarker tests behavior when no marker exists
func TestDetectProjectRootNoMarker(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(nestedDir)

	_, err := DetectProjectRoot()
	if err == nil {
		t.Errorf("Expected error when no marker found, got nil")
	}
}

// TestFindProjectRootMarker tests marker discovery
func TestFindProjectRootMarker(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git marker
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git: %v", err)
	}

	subDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create src: %v", err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(subDir)

	marker, path, err := FindProjectRootMarker()
	if err != nil {
		t.Fatalf("FindProjectRootMarker failed: %v", err)
	}

	if marker != ".git" {
		t.Errorf("Expected marker '.git', got %q", marker)
	}

	expectedPath := filepath.Join(tmpDir, ".git")
	if !pathsEqual(path, expectedPath) {
		t.Errorf("Expected path %q, got %q", expectedPath, path)
	}
}

// TestGetCrushConfigPathOrDefault tests fallback behavior
func TestGetCrushConfigPathOrDefault(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	fallback := "/etc/crush/config.json"

	// GetCrushConfigPath should succeed (returns CWD/.crush.json as fallback)
	// so GetCrushConfigPathOrDefault should return that path, not the fallback
	result := GetCrushConfigPathOrDefault(fallback)

	// The function returns CWD/.crush.json, not the fallback
	// because GetCrushConfigPath always succeeds
	// We need to resolve tmpDir for comparison (handles /var vs /private/var symlink on macOS)
	resolvedTmpDir := resolveSymlinks(tmpDir)
	expectedPath := filepath.Join(resolvedTmpDir, ".crush.json")
	if result != expectedPath {
		t.Errorf("Expected %q, got %q", expectedPath, result)
	}
}

// TestDetectProjectRootMarkerPriority tests marker priority (first found wins)
func TestDetectProjectRootMarkerPriority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple markers - .taskmaster should be found first based on the markers list
	for _, marker := range []string{".taskmaster", ".git", "go.mod"} {
		markerPath := filepath.Join(tmpDir, marker)
		if marker == "go.mod" {
			// go.mod is a file
			if err := os.WriteFile(markerPath, []byte("module test\n"), 0644); err != nil {
				t.Fatalf("Failed to create %s: %v", marker, err)
			}
		} else {
			// .taskmaster and .git are directories
			if err := os.MkdirAll(markerPath, 0755); err != nil {
				t.Fatalf("Failed to create %s: %v", marker, err)
			}
		}
	}

	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create sub: %v", err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(subDir)

	root, err := DetectProjectRoot()
	if err != nil {
		t.Fatalf("DetectProjectRoot failed: %v", err)
	}

	if !pathsEqual(root, tmpDir) {
		t.Errorf("Expected root %q, got %q", tmpDir, root)
	}
}

// resolveSymlinks resolves symlinks in a path for comparison purposes
func resolveSymlinks(p string) string {
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return p // Return original if resolution fails
	}
	return resolved
}

// pathsEqual compares two paths, handling symlink differences on macOS
func pathsEqual(a, b string) bool {
	return resolveSymlinks(a) == resolveSymlinks(b)
}

func TestCrushConfigMarshalUnmarshal(t *testing.T) {
	config := &CrushConfig{
		Schema:  "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-4", Provider: "openai"},
		},
		Version: "1.0",
		Options: CrushOptions{
			ContextPaths: []string{"README.md", "SPEC.md"},
			SkillsPaths:  []string{"./.crush/skills", "./.crush/custom"},
		},
		ExtraFields: make(map[string]interface{}),
	}

	// Marshal
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal
	loaded := &CrushConfig{}
	if err := json.Unmarshal(data, loaded); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify fields
	if loaded.Schema != config.Schema {
		t.Errorf("Schema mismatch: got %q, want %q", loaded.Schema, config.Schema)
	}
	
	// Verify models
	largeModel, ok := loaded.Models["large"]
	if !ok {
		t.Errorf("large model not found in loaded config")
	}
	if largeModel.Model != "gpt-4" {
		t.Errorf("Model mismatch: got %q, want %q", largeModel.Model, "gpt-4")
	}
	
	if len(loaded.Options.ContextPaths) != len(config.Options.ContextPaths) {
		t.Errorf("ContextPaths length mismatch: got %d, want %d", len(loaded.Options.ContextPaths), len(config.Options.ContextPaths))
	}
}

func TestCrushConfigPreserveUnknownFields(t *testing.T) {
	// Create JSON with unknown fields and the new models structure
	jsonData := `{
		"$schema": "https://charm.land/crush.json",
		"models": {
			"large": {
				"model": "gpt-4",
				"provider": "openai"
			}
		},
		"version": "1.0",
		"customField": "customValue",
		"nestedUnknown": {
			"key": "value"
		},
		"options": {
			"context_paths": ["README.md"],
			"skills_paths": ["./.crush/skills"],
			"unknownOption": "should be preserved"
		}
	}`

	// Unmarshal
	config := &CrushConfig{}
	if err := json.Unmarshal([]byte(jsonData), config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify known fields (models)
	largeModel, ok := config.Models["large"]
	if !ok {
		t.Errorf("large model not found")
	}
	if largeModel.Model != "gpt-4" {
		t.Errorf("Model mismatch: got %q, want %q", largeModel.Model, "gpt-4")
	}

	// Verify unknown fields are preserved
	if val, ok := config.ExtraFields["customField"]; !ok || val != "customValue" {
		t.Errorf("Unknown field not preserved: customField")
	}

	// Marshal back and verify unknown fields are in output
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if _, ok := unmarshaled["customField"]; !ok {
		t.Errorf("Unknown field not in marshaled output")
	}
}

func TestLoadCrushConfigNonexistent(t *testing.T) {
	// Temporarily change the config path function to use a temp directory
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Load config when file doesn't exist
	config, err := LoadCrushConfig()
	if err != nil {
		t.Fatalf("LoadCrushConfig failed: %v", err)
	}

	// Should return defaults
	if config == nil {
		t.Errorf("Expected default config, got nil")
	}
	if config.Schema != "https://charm.land/crush.json" {
		t.Errorf("Expected default schema, got %q", config.Schema)
	}
}

func TestLoadCrushConfigExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".crush.json")

	// Create a test config file with the new models structure
	testData := `{
		"$schema": "https://charm.land/crush.json",
		"models": {
			"large": {
				"model": "claude-3-sonnet",
				"provider": "anthropic"
			}
		},
		"version": "1.0",
		"options": {
			"context_paths": ["README.md"],
			"skills_paths": ["./.crush/skills"]
		}
	}`

	if err := os.WriteFile(configPath, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Temporarily change working directory and taskmaster location
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create a .taskmaster directory so findTaskMasterDir works
	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Load the config
	config, err := LoadCrushConfig()
	if err != nil {
		t.Fatalf("LoadCrushConfig failed: %v", err)
	}

	largeModel, ok := config.Models["large"]
	if !ok {
		t.Errorf("large model not found in config")
	}
	if largeModel.Model != "claude-3-sonnet" {
		t.Errorf("Model mismatch: got %q, want %q", largeModel.Model, "claude-3-sonnet")
	}

	if len(config.Options.ContextPaths) != 1 {
		t.Errorf("Expected 1 context path, got %d", len(config.Options.ContextPaths))
	}
}

func TestSaveCrushConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".crush.json")

	// Change to temp directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create .taskmaster so findTaskMasterDir works
	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	config := &CrushConfig{
		Schema:  "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-4-turbo", Provider: "openai"},
		},
		Version: "1.0",
		Options: CrushOptions{
			ContextPaths: []string{"README.md", "API.md"},
			SkillsPaths:  []string{"./.crush/skills"},
		},
		ExtraFields: make(map[string]interface{}),
	}

	// Save config
	if err := SaveCrushConfig(config); err != nil {
		t.Fatalf("SaveCrushConfig failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Load and verify
	loaded, err := LoadCrushConfig()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	largeModel, ok := loaded.Models["large"]
	if !ok {
		t.Errorf("large model not found in loaded config")
	}
	if largeModel.Model != "gpt-4-turbo" {
		t.Errorf("Model mismatch after save/load: got %q, want %q", largeModel.Model, "gpt-4-turbo")
	}
}

func TestSaveCrushConfigPreservesFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".crush.json")

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Create initial config with unknown fields and new models structure
	jsonData := `{
		"$schema": "https://charm.land/crush.json",
		"models": {
			"large": {
				"model": "gpt-4",
				"provider": "openai"
			}
		},
		"version": "1.0",
		"customField": "preserved",
		"options": {
			"context_paths": ["README.md"],
			"skills_paths": ["./.crush/skills"]
		}
	}`

	if err := os.WriteFile(configPath, []byte(jsonData), 0644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	// Load and modify
	config, err := LoadCrushConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Update the large model
	config.SetModelByType("large", "gpt-4-turbo", "openai")

	// Save
	if err := SaveCrushConfig(config); err != nil {
		t.Fatalf("SaveCrushConfig failed: %v", err)
	}

	// Verify the custom field is still there
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if val, ok := result["customField"]; !ok || val != "preserved" {
		t.Errorf("Custom field was not preserved after save")
	}

	// Verify model update
	if models, ok := result["models"].(map[string]interface{}); ok {
		if large, ok := models["large"].(map[string]interface{}); ok {
			if model, ok := large["model"].(string); !ok || model != "gpt-4-turbo" {
				t.Errorf("Model update was not saved correctly")
			}
		} else {
			t.Errorf("large model not found in saved config")
		}
	} else {
		t.Errorf("models field not found in saved config")
	}
}

func TestUpdateCrushModel(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Create initial config with the new models structure
	initialConfig := &CrushConfig{
		Schema: "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-3.5-turbo", Provider: "openai"},
		},
		Options: CrushOptions{
			ContextPaths: []string{"README.md"},
			SkillsPaths:  []string{"./.crush/skills"},
		},
		ExtraFields: make(map[string]interface{}),
	}

	if err := SaveCrushConfig(initialConfig); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Update model using new signature
	if err := UpdateCrushModel("claude-3-opus", "anthropic", "large"); err != nil {
		t.Fatalf("UpdateCrushModel failed: %v", err)
	}

	// Load and verify
	loaded, err := LoadCrushConfig()
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}

	largeModel, ok := loaded.Models["large"]
	if !ok {
		t.Errorf("large model not found after update")
	}
	if largeModel.Model != "claude-3-opus" {
		t.Errorf("Model not updated: got %q, want %q", largeModel.Model, "claude-3-opus")
	}
	if largeModel.Provider != "anthropic" {
		t.Errorf("Provider not updated: got %q, want %q", largeModel.Provider, "anthropic")
	}

	// Verify other fields are preserved
	if len(loaded.Options.ContextPaths) != 1 {
		t.Errorf("ContextPaths were lost during update")
	}
}

// TestUpdateCrushModelValidation tests input validation for UpdateCrushModel
func TestUpdateCrushModelValidation(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Test empty modelID
	err := UpdateCrushModel("", "openai", "large")
	if err == nil {
		t.Errorf("Expected error for empty modelID")
	}
	if err != nil && !strings.Contains(err.Error(), "modelID") {
		t.Errorf("Expected error message to mention modelID, got: %v", err)
	}

	// Test empty provider
	err = UpdateCrushModel("gpt-4", "", "large")
	if err == nil {
		t.Errorf("Expected error for empty provider")
	}
	if err != nil && !strings.Contains(err.Error(), "provider") {
		t.Errorf("Expected error message to mention provider, got: %v", err)
	}

	// Test empty modelType
	err = UpdateCrushModel("gpt-4", "openai", "")
	if err == nil {
		t.Errorf("Expected error for empty modelType")
	}
	if err != nil && !strings.Contains(err.Error(), "modelType") {
		t.Errorf("Expected error message to mention modelType, got: %v", err)
	}
}

func TestInitCrushConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".crush.json")

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Initialize - should create file with defaults
	config, err := InitCrushConfig()
	if err != nil {
		t.Fatalf("InitCrushConfig failed: %v", err)
	}

	if config == nil {
		t.Errorf("InitCrushConfig returned nil config")
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created by InitCrushConfig")
	}

	// Models should be initialized as empty map
	if config.Models == nil {
		t.Errorf("Models map should be initialized")
	}
	
	// Verify other model types can be compared after loading
	if len(config.Models) != 0 {
		t.Errorf("Default config should have empty models map")
	}

	// Initialize again - should load existing file
	config2, err := InitCrushConfig()
	if err != nil {
		t.Fatalf("InitCrushConfig second call failed: %v", err)
	}

	if len(config2.Models) != len(config.Models) {
		t.Errorf("Second init returned different config")
	}
}

func TestInitCrushConfigWithExistingFile(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Create existing config with new models structure
	existingConfig := &CrushConfig{
		Schema:      "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "existing-model", Provider: "test"},
		},
		Version:     "1.0",
		ExtraFields: make(map[string]interface{}),
	}

	if err := SaveCrushConfig(existingConfig); err != nil {
		t.Fatalf("Failed to save existing config: %v", err)
	}

	// Init should load existing
	config, err := InitCrushConfig()
	if err != nil {
		t.Fatalf("InitCrushConfig failed: %v", err)
	}

	largeModel, ok := config.Models["large"]
	if !ok {
		t.Errorf("large model not found in loaded config")
	}
	if largeModel.Model != "existing-model" {
		t.Errorf("InitCrushConfig did not load existing config: got model %q, want %q", largeModel.Model, "existing-model")
	}
}

func TestCrushConfigGettersSetters(t *testing.T) {
	config := getDefaultCrushConfig()

	// Test multi-model getter/setter
	config.SetModelByType("large", "new-model", "openai")
	model, ok := config.GetModelByType("large")
	if !ok {
		t.Errorf("Model getter failed: large model not found")
	}
	if model.Model != "new-model" {
		t.Errorf("Model getter/setter failed: got %q, want %q", model.Model, "new-model")
	}
	if model.Provider != "openai" {
		t.Errorf("Provider getter/setter failed: got %q, want %q", model.Provider, "openai")
	}

	// Test ContextPaths
	paths := []string{"file1.md", "file2.md"}
	config.SetContextPaths(paths)
	if len(config.GetContextPaths()) != 2 {
		t.Errorf("ContextPaths getter/setter failed")
	}

	// Test SkillsPaths
	skillPaths := []string{"./skills1", "./skills2"}
	config.SetSkillsPaths(skillPaths)
	if len(config.GetSkillsPaths()) != 2 {
		t.Errorf("SkillsPaths getter/setter failed")
	}
}

func TestCrushConfigAddRemoveContextPath(t *testing.T) {
	config := getDefaultCrushConfig()

	// Add paths
	config.AddContextPath("README.md")
	config.AddContextPath("SPEC.md")

	if len(config.GetContextPaths()) != 2 {
		t.Errorf("AddContextPath failed: expected 2 paths, got %d", len(config.GetContextPaths()))
	}

	// Adding duplicate should be no-op
	config.AddContextPath("README.md")
	if len(config.GetContextPaths()) != 2 {
		t.Errorf("AddContextPath allowed duplicate: expected 2 paths, got %d", len(config.GetContextPaths()))
	}

	// Remove path
	config.RemoveContextPath("README.md")
	if len(config.GetContextPaths()) != 1 {
		t.Errorf("RemoveContextPath failed: expected 1 path, got %d", len(config.GetContextPaths()))
	}

	if config.GetContextPaths()[0] != "SPEC.md" {
		t.Errorf("Wrong path remained after removal")
	}
}

func TestCrushConfigEmptyFields(t *testing.T) {
	config := &CrushConfig{
		ExtraFields: make(map[string]interface{}),
		Options: CrushOptions{
			ExtraOptions: make(map[string]interface{}),
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal empty config: %v", err)
	}

	// Verify omitempty works - empty fields should not appear
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Empty models should not be in JSON (omitempty should work)
	if _, ok := result["models"]; ok && result["models"] == nil {
		t.Errorf("Nil models field was included in JSON")
	}
}

func TestCrushConfigJSONFormatting(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	config := &CrushConfig{
		Schema: "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-4", Provider: "openai"},
		},
		Options: CrushOptions{
			ContextPaths: []string{"README.md"},
			SkillsPaths:  []string{"./.crush/skills"},
		},
		ExtraFields: make(map[string]interface{}),
	}

	if err := SaveCrushConfig(config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read raw file to check formatting
	data, err := os.ReadFile(filepath.Join(tmpDir, ".crush.json"))
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Check that it's indented (contains newlines and spaces)
	content := string(data)
	if !contains(content, "  ") {
		t.Errorf("Config file is not properly indented")
	}
}

func TestCrushConfigErrorHandling(t *testing.T) {
	// Test with non-existent directory for permission test
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Try to get config path - should work without proper .taskmaster
	path, err := GetCrushConfigPath()
	if err != nil {
		// This is acceptable - just means no project root found
		t.Logf("GetCrushConfigPath returned expected error when no .taskmaster found: %v", err)
	} else {
		// Should still have a path in current directory
		if path == "" {
			t.Errorf("GetCrushConfigPath returned empty path")
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkMarshalCrushConfig benchmarks config marshaling
func BenchmarkMarshalCrushConfig(b *testing.B) {
	config := &CrushConfig{
		Schema:  "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-4", Provider: "openai"},
		},
		Version: "1.0",
		Options: CrushOptions{
			ContextPaths: []string{"README.md", "API.md", "SPEC.md"},
			SkillsPaths:  []string{"./.crush/skills", "./.crush/custom"},
		},
		ExtraFields: map[string]interface{}{
			"custom1": "value1",
			"custom2": "value2",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(config)
	}
}

// TestSaveCrushConfigAtomic tests atomic write behavior
func TestSaveCrushConfigAtomic(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	config := &CrushConfig{
		Schema:      "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-4", Provider: "openai"},
		},
		ExtraFields: make(map[string]interface{}),
		Options: CrushOptions{
			ExtraOptions: make(map[string]interface{}),
		},
	}

	// Save config
	if err := SaveCrushConfig(config); err != nil {
		t.Fatalf("SaveCrushConfig failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".crush.json")

	// Verify file exists and is valid
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var loaded CrushConfig
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	largeModel, ok := loaded.Models["large"]
	if !ok {
		t.Errorf("large model not found after atomic write")
	}
	if largeModel.Model != "gpt-4" {
		t.Errorf("Model mismatch after atomic write")
	}

	// Test that no temporary files are left behind
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if entry.Name() != ".crush.json" && entry.Name() != ".taskmaster" {
			t.Errorf("Unexpected file left in directory: %s", entry.Name())
		}
	}
}

// TestSaveCrushConfigPreservesPermissions tests that existing file permissions are preserved
func TestSaveCrushConfigPreservesPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".crush.json")

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Create initial config with custom permissions
	initialConfig := &CrushConfig{
		Schema:      "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "initial", Provider: "test"},
		},
		ExtraFields: make(map[string]interface{}),
		Options: CrushOptions{
			ExtraOptions: make(map[string]interface{}),
		},
	}

	if err := SaveCrushConfig(initialConfig); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Set custom permissions
	if err := os.Chmod(configPath, 0600); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}

	// Update config
	initialConfig.SetModelByType("large", "updated", "test")
	if err := SaveCrushConfig(initialConfig); err != nil {
		t.Fatalf("Failed to save updated config: %v", err)
	}

	// Check that permissions were preserved
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config: %v", err)
	}

	if fileInfo.Mode().Perm() != 0600 {
		t.Errorf("Permissions not preserved: got %o, want 0600", fileInfo.Mode().Perm())
	}
}

// BenchmarkUnmarshalCrushConfig benchmarks config unmarshaling
func BenchmarkUnmarshalCrushConfig(b *testing.B) {
	jsonData := []byte(`{
		"$schema": "https://charm.land/crush.json",
		"models": {
			"large": {
				"model": "gpt-4",
				"provider": "openai"
			}
		},
		"version": "1.0",
		"options": {
			"context_paths": ["README.md", "API.md"],
			"skills_paths": ["./.crush/skills"]
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var config CrushConfig
		json.Unmarshal(jsonData, &config)
	}
}

// TestGetModelByType tests the GetModelByType helper method
func TestGetModelByType(t *testing.T) {
	config := &CrushConfig{
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-4", Provider: "openai"},
			"small": {Model: "gpt-3.5-turbo", Provider: "openai"},
		},
	}

	// Test retrieving existing model
	model, ok := config.GetModelByType("large")
	if !ok {
		t.Errorf("Expected to find large model")
	}
	if model.Model != "gpt-4" {
		t.Errorf("Expected gpt-4, got %s", model.Model)
	}
	if model.Provider != "openai" {
		t.Errorf("Expected openai provider, got %s", model.Provider)
	}

	// Test retrieving non-existing model
	_, ok = config.GetModelByType("medium")
	if ok {
		t.Errorf("Expected not to find medium model")
	}

	// Test with nil Models map
	configNil := &CrushConfig{}
	_, ok = configNil.GetModelByType("large")
	if ok {
		t.Errorf("Expected not to find model when Models is nil")
	}
}

// TestSetModelByType tests the SetModelByType helper method
func TestSetModelByType(t *testing.T) {
	// Test setting model on existing map
	config := &CrushConfig{
		Models: make(map[string]SelectedModel),
	}

	config.SetModelByType("large", "gpt-4", "openai")
	model, ok := config.Models["large"]
	if !ok {
		t.Errorf("Expected large model to be set")
	}
	if model.Model != "gpt-4" {
		t.Errorf("Expected gpt-4, got %s", model.Model)
	}
	if model.Provider != "openai" {
		t.Errorf("Expected openai, got %s", model.Provider)
	}

	// Test setting model when Models is nil (should initialize)
	configNil := &CrushConfig{}
	configNil.SetModelByType("small", "gpt-3.5-turbo", "openai")
	if configNil.Models == nil {
		t.Errorf("Expected Models map to be initialized")
	}
	model, ok = configNil.Models["small"]
	if !ok {
		t.Errorf("Expected small model to be set")
	}
	if model.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected gpt-3.5-turbo, got %s", model.Model)
	}

	// Test updating existing model
	config.SetModelByType("large", "gpt-4-turbo", "openai")
	model, ok = config.Models["large"]
	if !ok {
		t.Errorf("Expected large model to still exist")
	}
	if model.Model != "gpt-4-turbo" {
		t.Errorf("Expected gpt-4-turbo after update, got %s", model.Model)
	}
}

// TestRemoveModelByType tests the RemoveModelByType helper method
func TestRemoveModelByType(t *testing.T) {
	// Test removing existing model
	config := &CrushConfig{
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-4", Provider: "openai"},
			"small": {Model: "gpt-3.5-turbo", Provider: "openai"},
		},
	}

	config.RemoveModelByType("large")
	if _, ok := config.Models["large"]; ok {
		t.Errorf("Expected large model to be removed")
	}
	if _, ok := config.Models["small"]; !ok {
		t.Errorf("Expected small model to remain")
	}

	// Test removing non-existing model (should be no-op)
	config.RemoveModelByType("medium")
	if len(config.Models) != 1 {
		t.Errorf("Expected map to still have 1 entry")
	}

	// Test removing from nil Models map (should be no-op, no panic)
	configNil := &CrushConfig{}
	configNil.RemoveModelByType("large") // Should not panic
	if configNil.Models != nil {
		t.Errorf("Expected Models to remain nil")
	}
}

// TestMultiModelMarshalUnmarshal tests marshaling/unmarshaling with multiple models
func TestMultiModelMarshalUnmarshal(t *testing.T) {
	config := &CrushConfig{
		Schema: "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large":  {Model: "gpt-4", Provider: "openai"},
			"small":  {Model: "gpt-3.5-turbo", Provider: "openai"},
			"medium": {Model: "claude-3-sonnet", Provider: "anthropic"},
		},
		Version:     "1.0",
		ExtraFields: make(map[string]interface{}),
	}

	// Marshal
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	loaded := &CrushConfig{}
	if err := json.Unmarshal(data, loaded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify all models are preserved
	if len(loaded.Models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(loaded.Models))
	}

	largeModel, ok := loaded.Models["large"]
	if !ok || largeModel.Model != "gpt-4" {
		t.Errorf("large model not preserved correctly")
	}

	smallModel, ok := loaded.Models["small"]
	if !ok || smallModel.Model != "gpt-3.5-turbo" {
		t.Errorf("small model not preserved correctly")
	}

	mediumModel, ok := loaded.Models["medium"]
	if !ok || mediumModel.Model != "claude-3-sonnet" || mediumModel.Provider != "anthropic" {
		t.Errorf("medium model not preserved correctly")
	}
}

// TestUpdateCrushModelMultipleTypes tests updating different model types
func TestUpdateCrushModelMultipleTypes(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Create initial config
	initialConfig := &CrushConfig{
		Schema: "https://charm.land/crush.json",
		Models: map[string]SelectedModel{
			"large": {Model: "gpt-4", Provider: "openai"},
		},
		Options: CrushOptions{
			ContextPaths: []string{"README.md"},
			SkillsPaths:  []string{"./.crush/skills"},
		},
		ExtraFields: make(map[string]interface{}),
	}

	if err := SaveCrushConfig(initialConfig); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Update small model (new entry)
	if err := UpdateCrushModel("gpt-3.5-turbo", "openai", "small"); err != nil {
		t.Fatalf("UpdateCrushModel failed: %v", err)
	}

	// Load and verify
	loaded, err := LoadCrushConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify large model is unchanged
	largeModel, ok := loaded.Models["large"]
	if !ok || largeModel.Model != "gpt-4" {
		t.Errorf("large model should be unchanged")
	}

	// Verify small model was added
	smallModel, ok := loaded.Models["small"]
	if !ok {
		t.Errorf("small model not found")
	}
	if smallModel.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected gpt-3.5-turbo, got %s", smallModel.Model)
	}

	// Verify other fields preserved
	if len(loaded.Options.ContextPaths) != 1 {
		t.Errorf("ContextPaths were lost")
	}
}

// TestModelsOmitEmpty tests that empty Models map is omitted from JSON
func TestModelsOmitEmpty(t *testing.T) {
	config := &CrushConfig{
		Schema:      "https://charm.land/crush.json",
		Models:      make(map[string]SelectedModel), // Empty map
		Version:     "1.0",
		ExtraFields: make(map[string]interface{}),
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Empty models map should be omitted due to omitempty
	if _, ok := result["models"]; ok {
		t.Errorf("Empty models map should be omitted from JSON")
	}
}

// --- Migration Tests ---

// TestLegacyModelMigrationSimpleString tests migrating a simple string model field
func TestLegacyModelMigrationSimpleString(t *testing.T) {
	// Legacy format with just a model string
	jsonData := `{
		"$schema": "https://charm.land/crush.json",
		"model": "gpt-4",
		"version": "1.0",
		"options": {
			"context_paths": ["README.md"]
		}
	}`

	config := &CrushConfig{}
	if err := json.Unmarshal([]byte(jsonData), config); err != nil {
		t.Fatalf("Failed to unmarshal legacy config: %v", err)
	}

	// Verify migration to Models["large"]
	largeModel, ok := config.Models["large"]
	if !ok {
		t.Errorf("Expected legacy model to be migrated to Models['large']")
	}
	if largeModel.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %q", largeModel.Model)
	}
	if largeModel.Provider != "openai" {
		t.Errorf("Expected provider 'openai' (inferred), got %q", largeModel.Provider)
	}

	// Verify legacy "model" field is not in ExtraFields
	if _, ok := config.ExtraFields["model"]; ok {
		t.Errorf("Legacy 'model' field should not be in ExtraFields")
	}
}

// TestLegacyModelMigrationWithProvider tests migrating a model with provider info
func TestLegacyModelMigrationWithProvider(t *testing.T) {
	// Legacy format with model object containing provider
	jsonData := `{
		"$schema": "https://charm.land/crush.json",
		"model": {
			"model": "claude-3-opus",
			"provider": "anthropic"
		},
		"version": "1.0"
	}`

	config := &CrushConfig{}
	if err := json.Unmarshal([]byte(jsonData), config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	largeModel, ok := config.Models["large"]
	if !ok {
		t.Errorf("Expected legacy model to be migrated")
	}
	if largeModel.Model != "claude-3-opus" {
		t.Errorf("Expected model 'claude-3-opus', got %q", largeModel.Model)
	}
	if largeModel.Provider != "anthropic" {
		t.Errorf("Expected provider 'anthropic', got %q", largeModel.Provider)
	}
}

// TestLegacyModelMigrationPreservesUnknownFields tests that migration preserves other fields
func TestLegacyModelMigrationPreservesUnknownFields(t *testing.T) {
	jsonData := `{
		"$schema": "https://charm.land/crush.json",
		"model": "gpt-3.5-turbo",
		"version": "1.0",
		"customField": "customValue",
		"anotherField": {
			"nested": "data"
		}
	}`

	config := &CrushConfig{}
	if err := json.Unmarshal([]byte(jsonData), config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify migration happened
	if _, ok := config.Models["large"]; !ok {
		t.Errorf("Expected legacy model to be migrated")
	}

	// Verify unknown fields preserved
	if val, ok := config.ExtraFields["customField"]; !ok || val != "customValue" {
		t.Errorf("customField was not preserved")
	}

	if _, ok := config.ExtraFields["anotherField"]; !ok {
		t.Errorf("anotherField was not preserved")
	}

	// Verify legacy "model" field is NOT in ExtraFields
	if _, ok := config.ExtraFields["model"]; ok {
		t.Errorf("Legacy 'model' field should not be in ExtraFields")
	}
}

// TestLegacyModelMigrationIdempotent tests that re-loading a migrated config doesn't re-migrate
func TestLegacyModelMigrationIdempotent(t *testing.T) {
	// Start with legacy format
	jsonData := `{
		"model": "gpt-4"
	}`

	config := &CrushConfig{}
	if err := json.Unmarshal([]byte(jsonData), config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Marshal it back (should have new format)
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal again
	config2 := &CrushConfig{}
	if err := json.Unmarshal(data, config2); err != nil {
		t.Fatalf("Failed to unmarshal second time: %v", err)
	}

	// Verify it still has the model
	largeModel, ok := config2.Models["large"]
	if !ok {
		t.Errorf("Expected large model to persist")
	}
	if largeModel.Model != "gpt-4" {
		t.Errorf("Expected model to remain 'gpt-4', got %q", largeModel.Model)
	}

	// Verify the marshaled JSON doesn't contain the legacy "model" field
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if _, ok := result["model"]; ok {
		t.Errorf("Marshaled config should not contain legacy 'model' field")
	}

	// Should have "models" instead
	if _, ok := result["models"]; !ok {
		t.Errorf("Marshaled config should contain 'models' field")
	}
}

// TestLegacyModelMigrationWithExistingModels tests that migration doesn't override existing Models["large"]
func TestLegacyModelMigrationWithExistingModels(t *testing.T) {
	// Config with both legacy "model" and new "models" (user may have partially migrated)
	jsonData := `{
		"model": "gpt-3.5-turbo",
		"models": {
			"large": {
				"model": "claude-3-opus",
				"provider": "anthropic"
			}
		}
	}`

	config := &CrushConfig{}
	if err := json.Unmarshal([]byte(jsonData), config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Should keep the existing Models["large"], not overwrite with legacy
	largeModel, ok := config.Models["large"]
	if !ok {
		t.Errorf("Expected large model to exist")
	}
	if largeModel.Model != "claude-3-opus" {
		t.Errorf("Expected existing model 'claude-3-opus' to be preserved, got %q", largeModel.Model)
	}
	if largeModel.Provider != "anthropic" {
		t.Errorf("Expected existing provider 'anthropic', got %q", largeModel.Provider)
	}
}

// TestLegacyModelMigrationRoundTrip tests load-save-load preserves migration
func TestLegacyModelMigrationRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".crush.json")

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".taskmaster"), 0755); err != nil {
		t.Fatalf("Failed to create .taskmaster: %v", err)
	}

	// Write legacy format to file
	legacyData := `{
		"$schema": "https://charm.land/crush.json",
		"model": "gpt-4",
		"version": "1.0",
		"customField": "preserved"
	}`

	if err := os.WriteFile(configPath, []byte(legacyData), 0644); err != nil {
		t.Fatalf("Failed to write legacy config: %v", err)
	}

	// Load config (should trigger migration)
	config, err := LoadCrushConfig()
	if err != nil {
		t.Fatalf("LoadCrushConfig failed: %v", err)
	}

	// Verify migration happened
	largeModel, ok := config.Models["large"]
	if !ok {
		t.Errorf("Expected migration to Models['large']")
	}
	if largeModel.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %q", largeModel.Model)
	}

	// Save config (should write new format)
	if err := SaveCrushConfig(config); err != nil {
		t.Fatalf("SaveCrushConfig failed: %v", err)
	}

	// Read raw file and verify it has new format
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	// Should not have legacy "model" field
	if _, ok := result["model"]; ok {
		t.Errorf("Saved config should not have legacy 'model' field")
	}

	// Should have "models" with "large" entry
	if models, ok := result["models"].(map[string]interface{}); ok {
		if _, ok := models["large"]; !ok {
			t.Errorf("Saved config should have models.large")
		}
	} else {
		t.Errorf("Saved config should have 'models' field")
	}

	// Should preserve custom field
	if val, ok := result["customField"]; !ok || val != "preserved" {
		t.Errorf("Custom field was not preserved after save")
	}

	// Load again to verify idempotence
	config2, err := LoadCrushConfig()
	if err != nil {
		t.Fatalf("Second LoadCrushConfig failed: %v", err)
	}

	largeModel2, ok := config2.Models["large"]
	if !ok {
		t.Errorf("Expected large model to persist after second load")
	}
	if largeModel2.Model != "gpt-4" {
		t.Errorf("Expected model to remain 'gpt-4' after second load, got %q", largeModel2.Model)
	}
}

// TestProviderInference tests the inferProviderFromModel function
func TestProviderInference(t *testing.T) {
	tests := []struct {
		model    string
		provider string
	}{
		{"gpt-4", "openai"},
		{"gpt-3.5-turbo", "openai"},
		{"GPT-4-turbo", "openai"},
		{"claude-3-opus", "anthropic"},
		{"claude-2", "anthropic"},
		{"CLAUDE-3-Sonnet", "anthropic"},
		{"gemini-pro", "google"},
		{"palm-2", "google"},
		{"llama-2-70b", "meta"},
		{"llama-3", "meta"},
		{"mistral-7b", "mistral"},
		{"mixtral-8x7b", "mistral"},
		{"command-r-plus", "cohere"},
		{"unknown-model", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := inferProviderFromModel(tt.model)
			if result != tt.provider {
				t.Errorf("inferProviderFromModel(%q) = %q, want %q", tt.model, result, tt.provider)
			}
		})
	}
}

// TestLegacyModelMigrationEmptyModel tests that empty/invalid model doesn't cause issues
func TestLegacyModelMigrationEmptyModel(t *testing.T) {
	jsonData := `{
		"model": "",
		"version": "1.0"
	}`

	config := &CrushConfig{}
	if err := json.Unmarshal([]byte(jsonData), config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Should not create a Models["large"] entry for empty model
	if _, ok := config.Models["large"]; ok {
		t.Errorf("Should not migrate empty model string")
	}
}

// TestLegacyModelMigrationNullModel tests handling of null model value
func TestLegacyModelMigrationNullModel(t *testing.T) {
	jsonData := `{
		"model": null,
		"version": "1.0"
	}`

	config := &CrushConfig{}
	if err := json.Unmarshal([]byte(jsonData), config); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Should not create Models["large"] for null
	if _, ok := config.Models["large"]; ok {
		t.Errorf("Should not migrate null model")
	}
}

