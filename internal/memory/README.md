# Agent Memory for TaskMaster

This package provides a simple, local storage solution for LLM agent memory.

The current implementation uses a simple in-memory store with JSON file persistence, with a designed migration path to BadgerDB or other storage solutions as needed.

## Features

- Simple key-value storage for agent memory
- JSON serialization/deserialization
- Task activity logging
- README document storage and retrieval
- Easy migration path to other stores in future

## Usage for LLMs

The package provides a standalone command-line tool that can be executed by LLMs to retrieve and store memory:

```bash
# Store memory
memory store -key "readme:main" -file README.md
memory store -key "task:1.2" -value '{"title":"Implement auth","status":"pending"}' -json

# Retrieve memory
memory get -key "readme:main"  # Returns the README content
memory get -key "task:1.2"     # Returns the task data

# List keys
memory list  # List all keys
memory list -prefix "task:"  # List all task keys
memory list -json  # Return as JSON array

# Log task activity (useful for tracking progress)
memory log -task "1.2" -message "Started implementation of auth system"
memory log -task "1.2" -message "Completed JWT token generation"

# Get README files
memory readmes  # List all README files
```

## Usage in Go Code

```go
// Create memory store
helper, err := memory.DefaultHelper()
if err != nil {
    log.Fatalf("Failed to create memory helper: %v", err)
}
defer helper.Close()

// Store task information
taskInfo := map[string]string{
    "title": "Implement authentication",
    "status": "in_progress",
}
helper.StoreTaskInfo(ctx, "1.2", taskInfo)

// Retrieve task information
var retrievedInfo map[string]string
helper.GetTaskInfo(ctx, "1.2", &retrievedInfo)

// Log activity
helper.LogTaskActivity(ctx, "1.2", "Started implementation")

// Get logs
logs, _ := helper.GetTaskLogs(ctx, "1.2")

// Store README
helper.StoreReadme(ctx, "architecture", "# Architecture\n...")

// Get README
content, _ := helper.GetReadme(ctx, "architecture")
```

## Key Prefixes

The memory system uses key prefixes to organize different types of data:

- `task:` - Task information
- `readme:` - README documents
- `log:` - Task activity logs
- `context:` - Context information for LLMs

## Migration Path

The package defines a `Memory` interface that can be implemented by different storage backends:

```go
type Memory interface {
    Store(ctx context.Context, key string, value []byte) error
    Retrieve(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, prefix string) ([]string, error)
    Close() error
}
```

To switch to a different backend, you would:

1. Create a new implementation of the `Memory` interface
2. Pass it to `NewHelper()` instead of using `DefaultHelper()`

### Available Implementations

1. **InMemoryStorage** - The default implementation: simple in-memory map with JSON file persistence
   - Fast and simple
   - No dependencies
   - Single JSON file storage
   - Good for development and testing

2. **BadgerDB** - Future implementation: embedded key-value store
   - Faster for large datasets
   - Designed for performance
   - ACID transactions
   - More robust for production use
   
3. **Redis** (future) - For distributed deployment
   - Shared memory across instances
   - More complex setup (requires Redis server)
   - Advanced data structures
   - Scalability for larger deployments

## Building the Command Tool

```bash
go build -o bin/memory cmd/memory/main.go
```