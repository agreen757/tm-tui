package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInMemoryStorage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "memory-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new in-memory storage instance with file persistence
	memPath := filepath.Join(tempDir, "memory.json")
	mem, err := NewInMemoryStorage(memPath)
	if err != nil {
		t.Fatalf("Failed to create InMemoryStorage: %v", err)
	}
	defer mem.Close()

	ctx := context.Background()

	// Test storing and retrieving data
	key := "test-key"
	value := []byte("test-value")

	// Store data
	if err := mem.Store(ctx, key, value); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Retrieve data
	retrievedValue, err := mem.Retrieve(ctx, key)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}

	// Check if retrieved value matches
	if string(retrievedValue) != string(value) {
		t.Fatalf("Retrieved value doesn't match: got %s, want %s", retrievedValue, value)
	}

	// Test listing keys
	keys, err := mem.List(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}

	if len(keys) != 1 || keys[0] != key {
		t.Fatalf("List returned unexpected keys: %v", keys)
	}

	// Test deleting data
	if err := mem.Delete(ctx, key); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// Verify key no longer exists
	_, err = mem.Retrieve(ctx, key)
	if err != ErrKeyNotFound {
		t.Fatalf("Expected ErrKeyNotFound, got: %v", err)
	}
}

func TestHelper(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "helper-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new in-memory storage instance with file persistence
	memPath := filepath.Join(tempDir, "memory.json")
	mem, err := NewInMemoryStorage(memPath)
	if err != nil {
		t.Fatalf("Failed to create InMemoryStorage: %v", err)
	}

	// Create a new helper
	helper := NewHelper(mem)
	defer helper.Close()

	ctx := context.Background()

	// Test JSON storage and retrieval
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	testData := TestStruct{
		Name:  "test",
		Value: 42,
	}

	// Store JSON
	if err := helper.StoreJSON(ctx, "json-test", testData); err != nil {
		t.Fatalf("Failed to store JSON: %v", err)
	}

	// Retrieve JSON
	var retrievedData TestStruct
	if err := helper.RetrieveJSON(ctx, "json-test", &retrievedData); err != nil {
		t.Fatalf("Failed to retrieve JSON: %v", err)
	}

	// Check if retrieved data matches
	if retrievedData.Name != testData.Name || retrievedData.Value != testData.Value {
		t.Fatalf("Retrieved JSON doesn't match: got %+v, want %+v", retrievedData, testData)
	}

	// Test task info
	taskID := "1.2"
	taskInfo := map[string]interface{}{
		"title":       "Implement feature",
		"description": "Feature details",
		"status":      "pending",
	}

	if err := helper.StoreTaskInfo(ctx, taskID, taskInfo); err != nil {
		t.Fatalf("Failed to store task info: %v", err)
	}

	var retrievedTaskInfo map[string]interface{}
	if err := helper.GetTaskInfo(ctx, taskID, &retrievedTaskInfo); err != nil {
		t.Fatalf("Failed to get task info: %v", err)
	}

	// Check task info
	if retrievedTaskInfo["title"] != taskInfo["title"] {
		t.Fatalf("Retrieved task info doesn't match: got %+v, want %+v", retrievedTaskInfo, taskInfo)
	}

	// Test logging
	if err := helper.LogTaskActivity(ctx, taskID, "Started work"); err != nil {
		t.Fatalf("Failed to log activity: %v", err)
	}

	if err := helper.LogTaskActivity(ctx, taskID, "Completed work"); err != nil {
		t.Fatalf("Failed to log activity: %v", err)
	}

	logs, err := helper.GetTaskLogs(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 2 || logs[0] != "Started work" || logs[1] != "Completed work" {
		t.Fatalf("Retrieved logs don't match expected: %v", logs)
	}

	// Test README storage
	readmeName := "test-readme"
	readmeContent := "# Test README\nThis is a test."

	if err := helper.StoreReadme(ctx, readmeName, readmeContent); err != nil {
		t.Fatalf("Failed to store README: %v", err)
	}

	retrievedContent, err := helper.GetReadme(ctx, readmeName)
	if err != nil {
		t.Fatalf("Failed to get README: %v", err)
	}

	if retrievedContent != readmeContent {
		t.Fatalf("Retrieved README content doesn't match: got %s, want %s", retrievedContent, readmeContent)
	}

	readmes, err := helper.ListReadmes(ctx)
	if err != nil {
		t.Fatalf("Failed to list READMEs: %v", err)
	}

	if len(readmes) != 1 || readmes[0] != readmeName {
		t.Fatalf("Listed READMEs don't match expected: %v", readmes)
	}
}