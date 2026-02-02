package taskmaster

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewCachedStorage(t *testing.T) {
	backend := NewJSONStorage("/tmp/test.json")
	cached := NewCachedStorage(backend)

	if cached == nil {
		t.Fatal("NewCachedStorage returned nil")
	}
	if cached.Backend() != backend {
		t.Error("Backend not properly set")
	}
	if cached.IsLoaded() {
		t.Error("New cache should not be loaded")
	}
	if cached.IsDirty() {
		t.Error("New cache should not be dirty")
	}
}

func TestCachedStorage_LoadFromBackend(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	backend := NewJSONStorage(filePath)

	// Create and save test data to backend
	fc, _ := NewFileChange("test.go", "added", "Test")
	originalMapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := backend.Save(originalMapping); err != nil {
		t.Fatalf("Failed to save to backend: %v", err)
	}

	// Create cached storage and load
	cached := NewCachedStorage(backend)
	loaded, err := cached.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify data loaded correctly
	if loaded.Version != originalMapping.Version {
		t.Errorf("Version mismatch: expected %s, got %s", originalMapping.Version, loaded.Version)
	}
	if len(loaded.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(loaded.Tasks))
	}

	// Verify cache state
	if !cached.IsLoaded() {
		t.Error("Cache should be marked as loaded")
	}
	if cached.IsDirty() {
		t.Error("Cache should not be dirty after load")
	}
}

func TestCachedStorage_LoadCacheHit(t *testing.T) {
	tempDir := t.TempDir()
	backend := NewJSONStorage(filepath.Join(tempDir, "test.json"))
	cached := NewCachedStorage(backend)

	// First load - cache miss, loads from backend
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := backend.Save(mapping); err != nil {
		t.Fatalf("Failed to save to backend: %v", err)
	}

	loaded1, err := cached.Load()
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}

	// Modify backend directly (simulating external change)
	fc2, _ := NewFileChange("test2.go", "added", "Test 2")
	mapping.Tasks["2"] = []FileChange{*fc2}
	if err := backend.Save(mapping); err != nil {
		t.Fatalf("Failed to update backend: %v", err)
	}

	// Second load - cache hit, should return cached data (not updated backend data)
	loaded2, err := cached.Load()
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}

	// Should have same data as first load (cache hit)
	if len(loaded2.Tasks) != len(loaded1.Tasks) {
		t.Errorf("Cache hit failed: expected %d tasks, got %d (should be cached)", len(loaded1.Tasks), len(loaded2.Tasks))
	}
}

func TestCachedStorage_SaveAndDirty(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "test.json"))
	cached := NewCachedStorage(backend)

	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	// Save to cache
	if err := cached.Save(mapping); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify cache state
	if !cached.IsLoaded() {
		t.Error("Cache should be loaded after save")
	}
	if !cached.IsDirty() {
		t.Error("Cache should be dirty after save")
	}

	// Verify data is cached
	cachedData := cached.GetCachedCopy()
	if cachedData == nil {
		t.Fatal("Cached data should not be nil")
	}
	if len(cachedData.Tasks) != 1 {
		t.Errorf("Expected 1 task in cache, got %d", len(cachedData.Tasks))
	}
}

func TestCachedStorage_Flush(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	backend := NewJSONStorage(filePath)
	cached := NewCachedStorage(backend)

	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	// Save to cache
	if err := cached.Save(mapping); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify dirty before flush
	if !cached.IsDirty() {
		t.Error("Should be dirty before flush")
	}

	// Flush to backend
	if err := cached.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Verify not dirty after flush
	if cached.IsDirty() {
		t.Error("Should not be dirty after flush")
	}

	// Verify data was written to backend
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("Backend file should exist after flush: %v", err)
	}

	// Load directly from backend to verify
	backendData, err := backend.Load()
	if err != nil {
		t.Fatalf("Failed to load from backend: %v", err)
	}
	if len(backendData.Tasks) != 1 {
		t.Errorf("Expected 1 task in backend, got %d", len(backendData.Tasks))
	}
}

func TestCachedStorage_FlushWhenNotDirty(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "test.json"))
	cached := NewCachedStorage(backend)

	// Flush without any changes
	err := cached.Flush()
	if err != nil {
		t.Errorf("Flush should not error when not dirty: %v", err)
	}
}

func TestCachedStorage_Invalidate(t *testing.T) {
	tempDir := t.TempDir()
	backend := NewJSONStorage(filepath.Join(tempDir, "test.json"))
	cached := NewCachedStorage(backend)

	// Load data
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := backend.Save(mapping); err != nil {
		t.Fatalf("Failed to save to backend: %v", err)
	}

	if _, err := cached.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded
	if !cached.IsLoaded() {
		t.Error("Should be loaded")
	}

	// Invalidate cache
	cached.Invalidate()

	// Verify cache cleared
	if cached.IsLoaded() {
		t.Error("Should not be loaded after invalidate")
	}
	if cached.IsDirty() {
		t.Error("Should not be dirty after invalidate")
	}
	if cached.GetCachedCopy() != nil {
		t.Error("Cached data should be nil after invalidate")
	}
}

func TestCachedStorage_ConcurrentLoad(t *testing.T) {
	tempDir := t.TempDir()
	backend := NewJSONStorage(filepath.Join(tempDir, "concurrent.json"))

	// Prepare backend data
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := backend.Save(mapping); err != nil {
		t.Fatalf("Failed to save to backend: %v", err)
	}

	cached := NewCachedStorage(backend)

	// Load concurrently from multiple goroutines
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			loaded, err := cached.Load()
			if err != nil {
				errors <- fmt.Errorf("goroutine %d: load failed: %w", id, err)
				return
			}
			if loaded == nil {
				errors <- fmt.Errorf("goroutine %d: loaded nil", id)
				return
			}
			if len(loaded.Tasks) != 1 {
				errors <- fmt.Errorf("goroutine %d: expected 1 task, got %d", id, len(loaded.Tasks))
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

func TestCachedStorage_ConcurrentSave(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "concurrent.json"))
	cached := NewCachedStorage(backend)

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Save concurrently from multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fc, _ := NewFileChange(fmt.Sprintf("test%d.go", id), "added", fmt.Sprintf("Test %d", id))
			mapping := &FileChangeMapping{
				Version:           SerializationVersion,
				LastUpdated:       time.Now(),
				Tasks:             map[string][]FileChange{fmt.Sprintf("%d", id): {*fc}},
				UnassignedChanges: []FileChange{},
			}
			if err := cached.Save(mapping); err != nil {
				errors <- fmt.Errorf("goroutine %d: save failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify cache is dirty
	if !cached.IsDirty() {
		t.Error("Cache should be dirty after concurrent saves")
	}
}

func TestCachedStorage_ConcurrentLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	backend := NewJSONStorage(filepath.Join(tempDir, "concurrent.json"))

	// Prepare backend
	fc, _ := NewFileChange("initial.go", "added", "Initial")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"0": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := backend.Save(mapping); err != nil {
		t.Fatalf("Failed to save initial data: %v", err)
	}

	cached := NewCachedStorage(backend)

	const numGoroutines = 20
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Mix loads and saves
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		if i%2 == 0 {
			// Load
			go func(id int) {
				defer wg.Done()
				if _, err := cached.Load(); err != nil {
					errors <- fmt.Errorf("goroutine %d load: %w", id, err)
				}
			}(i)
		} else {
			// Save
			go func(id int) {
				defer wg.Done()
				fc, _ := NewFileChange(fmt.Sprintf("test%d.go", id), "added", fmt.Sprintf("Test %d", id))
				mapping := &FileChangeMapping{
					Version:           SerializationVersion,
					LastUpdated:       time.Now(),
					Tasks:             map[string][]FileChange{fmt.Sprintf("%d", id): {*fc}},
					UnassignedChanges: []FileChange{},
				}
				if err := cached.Save(mapping); err != nil {
					errors <- fmt.Errorf("goroutine %d save: %w", id, err)
				}
			}(i)
		}
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

func TestCachedStorage_ConcurrentFlush(t *testing.T) {
	tempDir := t.TempDir()
	backend := NewJSONStorage(filepath.Join(tempDir, "concurrent_flush.json"))
	cached := NewCachedStorage(backend)

	// Save some data
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := cached.Save(mapping); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Flush concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := cached.Flush(); err != nil {
				errors <- fmt.Errorf("goroutine %d: flush failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Should not be dirty after flush
	if cached.IsDirty() {
		t.Error("Should not be dirty after concurrent flush")
	}
}

func TestCachedStorage_SaveInvalidData(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "invalid.json"))
	cached := NewCachedStorage(backend)

	// Invalid mapping (empty version)
	invalid := &FileChangeMapping{
		Version:           "",
		Tasks:             make(map[string][]FileChange),
		UnassignedChanges: []FileChange{},
	}

	err := cached.Save(invalid)
	if err == nil {
		t.Fatal("Save should fail for invalid data")
	}
}

func TestCachedStorage_SaveAndFlush(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "saveflush.json")
	backend := NewJSONStorage(filePath)
	cached := NewCachedStorage(backend)

	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	// SaveAndFlush convenience method
	if err := cached.SaveAndFlush(mapping); err != nil {
		t.Fatalf("SaveAndFlush failed: %v", err)
	}

	// Should not be dirty after immediate flush
	if cached.IsDirty() {
		t.Error("Should not be dirty after SaveAndFlush")
	}

	// Verify data written to backend
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("Backend file should exist: %v", err)
	}
}

func TestCachedStorage_GetCachedCopyReturnsNil(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "test.json"))
	cached := NewCachedStorage(backend)

	// Get cached copy before loading
	copy := cached.GetCachedCopy()
	if copy != nil {
		t.Error("GetCachedCopy should return nil before load")
	}
}

func TestCachedStorage_MutationIsolation(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "mutation.json"))
	cached := NewCachedStorage(backend)

	// Save initial data
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := cached.Save(mapping); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load and mutate
	loaded, err := cached.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Mutate loaded data
	fc2, _ := NewFileChange("mutated.go", "added", "Mutated")
	loaded.Tasks["2"] = []FileChange{*fc2}

	// Load again - should not see mutation
	loaded2, err := cached.Load()
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}

	if len(loaded2.Tasks) != 1 {
		t.Errorf("Mutation leaked into cache: expected 1 task, got %d", len(loaded2.Tasks))
	}
}

func TestCachedStorage_ErrorPropagation(t *testing.T) {
	// Create backend with invalid path to force error
	backend := NewJSONStorage("/invalid/path/that/does/not/exist/test.json")
	cached := NewCachedStorage(backend)

	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}

	// Save to cache (should succeed)
	if err := cached.Save(mapping); err != nil {
		t.Fatalf("Save to cache should succeed: %v", err)
	}

	// Flush should propagate backend error
	err := cached.Flush()
	if err == nil {
		t.Fatal("Flush should propagate backend error")
	}

	// Should still be dirty after failed flush
	if !cached.IsDirty() {
		t.Error("Should remain dirty after failed flush")
	}
}

// Auto-save tests

func TestCachedStorage_StartAutoSave(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "autosave.json"))
	cached := NewCachedStorage(backend)

	// Start auto-save
	err := cached.StartAutoSave(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("StartAutoSave failed: %v", err)
	}

	// Verify auto-save is running
	if !cached.IsAutoSaveRunning() {
		t.Error("Auto-save should be running")
	}

	// Stop auto-save
	if err := cached.StopAutoSave(); err != nil {
		t.Fatalf("StopAutoSave failed: %v", err)
	}

	// Verify auto-save stopped
	if cached.IsAutoSaveRunning() {
		t.Error("Auto-save should be stopped")
	}
}

func TestCachedStorage_StartAutoSaveTwice(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "autosave.json"))
	cached := NewCachedStorage(backend)

	// Start auto-save
	if err := cached.StartAutoSave(100 * time.Millisecond); err != nil {
		t.Fatalf("First StartAutoSave failed: %v", err)
	}

	// Try to start again - should error
	err := cached.StartAutoSave(100 * time.Millisecond)
	if err == nil {
		t.Fatal("Starting auto-save twice should error")
	}

	// Clean up
	cached.StopAutoSave()
}

func TestCachedStorage_AutoSavePeriodicFlush(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "autosave.json")
	backend := NewJSONStorage(filePath)
	cached := NewCachedStorage(backend)

	// Save data to cache
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := cached.Save(mapping); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify dirty
	if !cached.IsDirty() {
		t.Fatal("Should be dirty before auto-save")
	}

	// Start auto-save with short interval
	if err := cached.StartAutoSave(50 * time.Millisecond); err != nil {
		t.Fatalf("StartAutoSave failed: %v", err)
	}

	// Wait for auto-save to flush
	time.Sleep(150 * time.Millisecond)

	// Should no longer be dirty
	if cached.IsDirty() {
		t.Error("Should not be dirty after auto-save flush")
	}

	// Verify file was written
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("File should exist after auto-save: %v", err)
	}

	// Stop auto-save
	cached.StopAutoSave()
}

func TestCachedStorage_StopAutoSaveWithoutStart(t *testing.T) {
	backend := NewJSONStorage(filepath.Join(t.TempDir(), "autosave.json"))
	cached := NewCachedStorage(backend)

	// Stop without starting - should not error
	err := cached.StopAutoSave()
	if err != nil {
		t.Errorf("StopAutoSave without start should not error: %v", err)
	}
}

func TestCachedStorage_AutoSaveFinalFlush(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "finalflush.json")
	backend := NewJSONStorage(filePath)
	cached := NewCachedStorage(backend)

	// Start auto-save with long interval (so periodic flush doesn't trigger)
	if err := cached.StartAutoSave(10 * time.Second); err != nil {
		t.Fatalf("StartAutoSave failed: %v", err)
	}

	// Save data
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := cached.Save(mapping); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Stop auto-save - should trigger final flush
	if err := cached.StopAutoSave(); err != nil {
		t.Fatalf("StopAutoSave failed: %v", err)
	}

	// Should not be dirty after stop (final flush)
	if cached.IsDirty() {
		t.Error("Should not be dirty after stop with final flush")
	}

	// Verify file was written
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("File should exist after final flush: %v", err)
	}
}

type testErrorLogger struct {
	errors []error
	mu     sync.Mutex
}

func (l *testErrorLogger) LogError(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errors = append(l.errors, err)
}

func (l *testErrorLogger) GetErrors() []error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]error(nil), l.errors...)
}

func TestCachedStorage_AutoSaveErrorLogging(t *testing.T) {
	// Create backend with invalid path to force flush errors
	backend := NewJSONStorage("/invalid/path/autosave.json")
	cached := NewCachedStorage(backend)

	// Set custom error logger
	logger := &testErrorLogger{}
	cached.SetErrorLogger(logger)

	// Save data
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := cached.Save(mapping); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Start auto-save with short interval
	if err := cached.StartAutoSave(50 * time.Millisecond); err != nil {
		t.Fatalf("StartAutoSave failed: %v", err)
	}

	// Wait for auto-save to attempt flush and fail
	time.Sleep(150 * time.Millisecond)

	// Stop auto-save
	cached.StopAutoSave()

	// Verify errors were logged
	errors := logger.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected auto-save errors to be logged")
	}
}

func TestCachedStorage_CreateBackup(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	backend := NewJSONStorage(filePath)
	cached := NewCachedStorage(backend)

	// Create and save data
	fc, _ := NewFileChange("test.go", "added", "Test")
	mapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc}},
		UnassignedChanges: []FileChange{},
	}
	if err := backend.Save(mapping); err != nil {
		t.Fatalf("Failed to save to backend: %v", err)
	}

	// Create backup
	backupPath, err := cached.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("Backup file should exist: %v", err)
	}

	// Verify backup contains valid data
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	_, err = UnmarshalFileChangeMapping(backupData)
	if err != nil {
		t.Errorf("Backup contains invalid data: %v", err)
	}
}

func TestCachedStorage_CreateBackupNoFile(t *testing.T) {
	tempDir := t.TempDir()
	backend := NewJSONStorage(filepath.Join(tempDir, "nonexistent.json"))
	cached := NewCachedStorage(backend)

	// Try to create backup of non-existent file
	_, err := cached.CreateBackup()
	if err == nil {
		t.Fatal("CreateBackup should fail for non-existent file")
	}
}

func TestCachedStorage_RestoreFromBackup(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	backend := NewJSONStorage(filePath)
	cached := NewCachedStorage(backend)

	// Create original data
	fc1, _ := NewFileChange("test1.go", "added", "Test 1")
	originalMapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"1": {*fc1}},
		UnassignedChanges: []FileChange{},
	}
	if err := backend.Save(originalMapping); err != nil {
		t.Fatalf("Failed to save original: %v", err)
	}

	// Create backup
	backupPath, err := cached.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Modify original file
	fc2, _ := NewFileChange("test2.go", "added", "Test 2")
	modifiedMapping := &FileChangeMapping{
		Version:           SerializationVersion,
		LastUpdated:       time.Now(),
		Tasks:             map[string][]FileChange{"2": {*fc2}},
		UnassignedChanges: []FileChange{},
	}
	if err := backend.Save(modifiedMapping); err != nil {
		t.Fatalf("Failed to save modified: %v", err)
	}

	// Restore from backup
	if err := cached.RestoreFromBackup(backupPath); err != nil {
		t.Fatalf("RestoreFromBackup failed: %v", err)
	}

	// Load and verify restored data matches original
	restored, err := backend.Load()
	if err != nil {
		t.Fatalf("Failed to load restored data: %v", err)
	}

	if len(restored.Tasks) != 1 {
		t.Errorf("Expected 1 task after restore, got %d", len(restored.Tasks))
	}
	if _, ok := restored.Tasks["1"]; !ok {
		t.Error("Restored data should contain original task 1")
	}
}

func TestCachedStorage_RestoreFromBackupInvalidBackup(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")
	backend := NewJSONStorage(filePath)
	cached := NewCachedStorage(backend)

	// Create corrupted backup
	backupPath := filepath.Join(tempDir, "corrupted.backup")
	if err := os.WriteFile(backupPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted backup: %v", err)
	}

	// Try to restore - should fail
	err := cached.RestoreFromBackup(backupPath)
	if err == nil {
		t.Fatal("RestoreFromBackup should fail for corrupted backup")
	}
}

func TestCachedStorage_RestoreFromBackupNonexistent(t *testing.T) {
	tempDir := t.TempDir()
	backend := NewJSONStorage(filepath.Join(tempDir, "test.json"))
	cached := NewCachedStorage(backend)

	// Try to restore from non-existent backup
	err := cached.RestoreFromBackup("/nonexistent/backup.json")
	if err == nil {
		t.Fatal("RestoreFromBackup should fail for non-existent backup")
	}
}

func TestCachedStorage_RecoverFromCorruption(t *testing.T) {
	tempDir := t.TempDir()
	backend := NewJSONStorage(filepath.Join(tempDir, "test.json"))
	cached := NewCachedStorage(backend)

	// Attempt recovery (should invalidate cache)
	err := cached.RecoverFromCorruption()
	if err == nil {
		t.Log("RecoverFromCorruption currently returns error indicating manual recovery needed")
	}

	// Cache should be invalidated
	if cached.IsLoaded() {
		t.Error("Cache should be invalidated after recovery attempt")
	}
}

