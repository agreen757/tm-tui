#!/bin/bash

echo "=== Test File Change Tracker ==="
echo

# Run backfill
echo "1. Running backfill..."
go run backfill_file_changes.go > /dev/null 2>&1
BEFORE_COUNT=$(grep -c '"1.1":' .taskmaster/file-changes.json)
echo "   Tasks with file changes (before): $BEFORE_COUNT"

# Check initial data
echo
echo "2. Sample backfilled data for task 1.1:"
grep -A 3 '"1.1":' .taskmaster/file-changes.json | head -5

# Start TUI in background
echo
echo "3. Starting TUI to test initialization..."
timeout 10s ./bin/tm-tui-fixed --tag file-change-tracker > /dev/null 2>&1 &
TUI_PID=$!

# Wait for initialization
sleep 3

# Check if data is preserved
AFTER_INIT_COUNT=$(grep -c '"1.1":' .taskmaster/file-changes.json)
echo "   Tasks with file changes (after init): $AFTER_INIT_COUNT"

# Wait for at least one refresh cycle (30 seconds)
echo
echo "4. Waiting 35 seconds for refresh cycle..."
sleep 35

# Check after refresh
AFTER_REFRESH_COUNT=$(grep -c '"1.1":' .taskmaster/file-changes.json)
echo "   Tasks with file changes (after refresh): $AFTER_REFRESH_COUNT"

# Kill TUI
kill $TUI_PID 2>/dev/null
wait $TUI_PID 2>/dev/null

echo
echo "=== Test Results ==="
echo "Before TUI:        $BEFORE_COUNT tasks"
echo "After init:        $AFTER_INIT_COUNT tasks"
echo "After refresh:     $AFTER_REFRESH_COUNT tasks"

if [ "$BEFORE_COUNT" = "$AFTER_INIT_COUNT" ] && [ "$BEFORE_COUNT" = "$AFTER_REFRESH_COUNT" ]; then
    echo
    echo "✅ SUCCESS: File changes preserved through init and refresh!"
    exit 0
else
    echo
    echo "❌ FAILED: File changes were lost"
    exit 1
fi
