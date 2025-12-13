# Memory Tool for LLM Agents

This command-line tool provides a simple interface for LLM agents to store and retrieve memory.

## Installation

```bash
# From the project root
go build -o $GOPATH/bin/memory cmd/memory/main.go

# Or install directly
go install github.com/adriangreen/tm-tui/cmd/memory@latest
```

## Usage

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

# Log task activity
memory log -task "1.2" -message "Started implementation of auth system"
memory log -task "1.2" -message "Completed JWT token generation"

# List README files
memory readmes
```

## Examples for LLM Use

When an LLM needs to remember context between runs, it can store and retrieve that information:

```bash
# Store information about a task
memory store -key "task:1.2" -value "This task involves implementing user authentication with JWT tokens." 

# Later, retrieve that information
memory get -key "task:1.2"

# Store logs about progress
memory log -task "1.2" -message "Started implementing JWT token generation"
memory log -task "1.2" -message "Completed user authentication flow"

# Store a README for future reference
memory store -key "readme:architecture" -file ARCHITECTURE.md

# List all stored memories with a specific prefix
memory list -prefix "task:"
```

## Storage Location

By default, memory is stored in `.taskmaster/memory.json` in the current working directory. This provides persistence between runs while maintaining a simple implementation.

## Migrating to Other Backends

The system is designed to support other storage backends in the future:

1. **Local Usage** (current): Simple JSON file-based storage
2. **Development**: BadgerDB embedded storage (planned)
3. **Production**: Redis or other distributed storage (future)

The command line interface will remain the same regardless of the storage backend used.