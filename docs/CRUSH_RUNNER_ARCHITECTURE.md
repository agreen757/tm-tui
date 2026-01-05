# Crush Runner Architecture & Performance

## Overview

The Task Runner integration executes Crush CLI commands as subprocesses and streams their output to the TUI in real-time. This document describes the architecture, performance considerations, and solutions to common issues.

## Architecture

### Message Flow

```
┌─────────────────┐
│  User Action    │
│  (Ctrl+R)       │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────────┐
│  StartCrushExecution (crush_runner.go:168)  │
│  - Validates crush binary                   │
│  - Returns TaskStartedMsg                   │
└────────┬────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────┐
│  ExecuteCrushSubprocess (crush_runner.go:191)│
│  - Creates buffered channel (1000 msgs)     │
│  - Spawns goroutine for subprocess          │
│  - Returns CrushExecutionSub immediately    │
└────────┬────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────┐
│  runCrushProcess (goroutine)                │
│  - Starts crush subprocess                  │
│  - Creates stdout/stderr scanners           │
│  - Spawns 2 goroutines for streaming        │
└────────┬────────────────────────────────────┘
         │
         ├──────────────┬──────────────┐
         ▼              ▼              ▼
    ┌────────┐    ┌────────┐    ┌──────────┐
    │ stdout │    │ stderr │    │ Process  │
    │ stream │    │ stream │    │  Wait    │
    └───┬────┘    └───┬────┘    └─────┬────┘
        │             │               │
        │ TaskOutputMsg                │
        │             │               │
        └─────────────┴───────────────┘
                      │
                      ▼
              ┌──────────────┐
              │   outCh      │
              │  (buffered)  │
              └──────┬───────┘
                     │
                     ▼
              ┌──────────────────────┐
              │  WaitForCrushMsg     │
              │  (app.go message     │
              │   subscription)      │
              └──────┬───────────────┘
                     │
                     ▼
              ┌──────────────────────┐
              │  TaskRunnerModal     │
              │  (UI update)         │
              └──────────────────────┘
```

## Performance Characteristics

### Channel Buffering

**Buffer Size**: 1000 messages (increased from 100)
- **Rationale**: Crush can produce output at ~10-50 lines/second during active execution
- **Headroom**: 1000-message buffer provides ~20-50 seconds of buffering for slow consumers
- **Memory**: Each message ~100-500 bytes = ~100-500KB total buffer memory

### Non-Blocking Sends

**Timeout Strategy**: 100ms timeout on channel sends
- **Why**: Prevents goroutines from blocking indefinitely if TUI is slow
- **Fallback**: Messages are always written to log file, UI may drop messages
- **Trade-off**: Complete logs vs real-time UI updates

### Scanner Buffer

**Buffer Size**: 1MB maximum line length
- **Rationale**: AI output can include large code blocks or JSON responses
- **Default**: Go's default scanner buffer is 64KB, which may truncate long lines
- **Safety**: Prevents "token too long" scanner errors

## Common Issues & Solutions

### Issue 1: TLS Connection Errors

**Symptom**:
```
ERROR: Agent processing failed: failed to start agent processing stream: 
Post "https://api.anthropic.com/v1/messages": remote error: tls: bad record MAC.
```

**Root Cause**: 
- Network connectivity issues
- API rate limiting
- Connection resets from Anthropic's servers
- NOT related to TUI buffering

**Solution**:
1. Check network connectivity: `curl -I https://api.anthropic.com`
2. Verify API key is valid
3. Check for rate limit headers in previous requests
4. Consider implementing exponential backoff in Crush CLI (upstream fix)

**Workaround**: Retry the task execution after a brief delay

### Issue 2: Slow Consumer Warnings in Crush Logs

**Symptom**:
```json
{"level":"WARN","msg":"message dropped due to slow consumer","name":"sessions"}
{"level":"WARN","msg":"message dropped due to slow consumer","name":"messages"}
```

**Root Cause**: 
- These warnings are from **Crush's internal event bus**, not the TUI integration
- Crush has multiple internal subscribers (sessions, messages, tools)
- When Crush's own goroutines can't keep up, it drops internal events
- This is a performance characteristic of Crush itself

**Impact on TUI**: 
- **None** - These are internal to Crush and don't affect the TUI's message channel
- TUI receives stdout/stderr from Crush's subprocess, not Crush's internal events

**Solution**: 
- No action needed in TUI
- Upstream Crush issue - consider reporting to Charm if it affects functionality

### Issue 3: Channel Backpressure

**Symptom**: Task execution appears to pause or slow down during high output

**Root Cause**: 
- TUI processes messages synchronously (one at a time)
- When output rate exceeds UI update rate, channel fills up
- Before fix: goroutines would block waiting for channel to drain
- After fix: messages timeout after 100ms and are dropped from UI (but logged)

**Current Behavior**:
- Complete output always written to log file
- UI shows most messages, may drop some during extreme bursts
- No goroutine blocking or process hangs

**Trade-offs**:
- **Pro**: Task never blocks, always makes progress
- **Con**: UI may not show every single line during bursts
- **Mitigation**: Log file has complete output for review

## Configuration

### Tuning Parameters

Located in `crush_runner.go`:

```go
// Channel buffer size
outCh := make(chan tea.Msg, 1000)

// Channel send timeout
case <-time.After(100 * time.Millisecond):

// Scanner buffer size
scanner.Buffer(buf, 1024*1024) // 1MB
```

**When to increase buffer**:
- Tasks consistently produce >1000 lines in <20 seconds
- Log shows many "Output channel full" warnings
- UI is missing critical output

**When to increase timeout**:
- UI update cycle is slower than 100ms (very slow terminals)
- Tasks are incorrectly reported as "dropping messages"

**When to increase scanner buffer**:
- Tasks output very long lines (>1MB)
- Scanner errors: "token too long"

## Debugging

### Enable Verbose Logging

Check the task-specific log file:
```bash
tail -f .taskmaster/logs/crush-run-<task-id>-<timestamp>.log
```

Or with tag context:
```bash
tail -f .taskmaster/<tag-name>/<task-id>.log
```

### Check for Dropped Messages

Search log for warnings:
```bash
grep "Output channel full" .taskmaster/logs/*.log
```

### Monitor Channel Performance

Add instrumentation (development only):
```go
// In runCrushProcess, after creating channel:
go func() {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        log.Printf("Channel buffer: %d/%d", len(outCh), cap(outCh))
    }
}()
```

## Future Improvements

### Potential Enhancements

1. **Adaptive buffering**: Increase buffer size dynamically based on load
2. **Priority messages**: Ensure errors/warnings always reach UI
3. **Rate limiting**: Throttle message sends to match UI refresh rate
4. **Batch updates**: Group multiple lines into single UI update
5. **Streaming compression**: Delta encoding for repeated output patterns

### Upstream Issues

1. **Crush TLS retry**: Add exponential backoff for API connection errors
2. **Crush event bus**: Investigate slow consumer warnings in Crush core
3. **Structured output**: JSON-formatted output for easier parsing

## Testing

### Load Testing

Test with high-output tasks:
```bash
# Generate 10,000 lines rapidly
crush run "Write a script that prints 10,000 lines of output"
```

### Slow Consumer Simulation

Artificially slow down UI updates:
```go
// In TaskRunnerModal.Update
case TaskOutputMsg:
    time.Sleep(200 * time.Millisecond) // Simulate slow UI
    tab.AddOutputLine(msg.Output)
```

Expected: Messages logged, some dropped from UI, no blocking

## References

- [Crush CLI](https://github.com/charmbracelet/crush)
- [Go Channels Best Practices](https://go.dev/blog/pipelines)
- [Bubble Tea Event Loop](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
