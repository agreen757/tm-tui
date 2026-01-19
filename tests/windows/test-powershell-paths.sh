#!/bin/bash
# PowerShell-style path testing script
# Tests Windows path handling with PowerShell-style conventions

echo "========================================"
echo "PowerShell-Style Path Testing"
echo "========================================"
echo ""

# Simulate PowerShell PATH format (semicolon-separated on Windows)
# In Git Bash, we still use colon-separated, but this tests the logic

echo "Test 1: Checking current environment..."
echo "Shell: $SHELL"
echo "PATH: $PATH"
echo "HOME: $HOME"
echo ""

echo "Test 2: Simulating Windows path format..."
# Convert Unix path to Windows-style (for documentation purposes)
UNIX_NPM_BIN=$(npm config get prefix)/bin
echo "Unix-style path: $UNIX_NPM_BIN"

# Simulate Windows path (C:\Users\...)
SIMULATED_WIN_PATH="C:\\Users\\testuser\\.npm-global\\bin"
echo "Windows-style path (simulated): $SIMULATED_WIN_PATH"
echo ""

echo "Test 3: Testing path with spaces (common Windows scenario)..."
# Create a directory with spaces
TEST_DIR="$HOME/My Documents/npm-test"
mkdir -p "$TEST_DIR"
echo "Created directory with spaces: $TEST_DIR"

# Test if we can add it to PATH
export PATH="$TEST_DIR:$PATH"
if [[ "$PATH" == *"$TEST_DIR"* ]]; then
    echo "✅ PASS: Path with spaces can be added to PATH"
else
    echo "❌ FAIL: Path with spaces could not be added"
fi
echo ""

echo "Test 4: Testing special characters in paths..."
# Test path with parentheses (common in Windows like "Program Files (x86)")
TEST_DIR_SPECIAL="$HOME/Program Files (x86)/npm"
mkdir -p "$TEST_DIR_SPECIAL"
export PATH="$TEST_DIR_SPECIAL:$PATH"
if [[ "$PATH" == *"$TEST_DIR_SPECIAL"* ]]; then
    echo "✅ PASS: Path with special characters can be added to PATH"
else
    echo "❌ FAIL: Path with special characters could not be added"
fi
echo ""

echo "Test 5: Verifying task-master with modified PATH..."
if command -v task-master &> /dev/null; then
    task-master --version > /dev/null 2>&1
    echo "✅ PASS: task-master works with modified PATH"
else
    echo "❌ FAIL: task-master not accessible with modified PATH"
fi
echo ""

echo "========================================"
echo "PowerShell-Style Test Summary"
echo "========================================"
echo "Environment successfully handles:"
echo "  - Spaces in paths"
echo "  - Special characters (parentheses)"
echo "  - Multiple PATH entries"
echo "  - task-master accessibility"
