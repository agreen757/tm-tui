#!/bin/bash
# Test script for Windows/Git Bash installation workflow

set -e  # Exit on error

echo "=================================="
echo "Windows Installation Test Script"
echo "=================================="
echo ""

# Function to print test results
print_test() {
    if [ $1 -eq 0 ]; then
        echo "✅ PASS: $2"
    else
        echo "❌ FAIL: $2"
        return 1
    fi
}

# Test 1: Check Node.js and npm are installed
echo "Test 1: Checking Node.js and npm installation..."
node --version
npm --version
print_test $? "Node.js and npm are installed"
echo ""

# Test 2: Check make is available
echo "Test 2: Checking make availability..."
make --version > /dev/null 2>&1
print_test $? "Make is available"
echo ""

# Test 3: Verify npm global directory is configured
echo "Test 3: Checking npm global directory configuration..."
NPM_PREFIX=$(npm config get prefix)
echo "npm prefix: $NPM_PREFIX"
if [[ "$NPM_PREFIX" == *".npm-global"* ]]; then
    print_test 0 "npm global directory is configured"
else
    print_test 1 "npm global directory is NOT configured"
fi
echo ""

# Test 4: Check if npm bin directory is in PATH
echo "Test 4: Checking if npm bin directory is in PATH..."
echo "Current PATH: $PATH"
if [[ "$PATH" == *".npm-global/bin"* ]]; then
    print_test 0 "npm bin directory is in PATH"
else
    print_test 1 "npm bin directory is NOT in PATH"
fi
echo ""

# Test 5: Install task-master-ai via Makefile
echo "Test 5: Installing task-master-ai..."
cd /home/testuser/workspace
if [ -f "Makefile" ]; then
    make check-task-master || true  # Don't fail if not installed yet
    echo "Running: make install-task-master"
    make install-task-master
    print_test $? "task-master installation via Makefile"
else
    echo "❌ Makefile not found in workspace"
    exit 1
fi
echo ""

# Test 6: Verify task-master is accessible
echo "Test 6: Verifying task-master is accessible..."
which task-master || true
if command -v task-master &> /dev/null; then
    print_test 0 "task-master is accessible in PATH"
    echo "task-master location: $(which task-master)"
else
    print_test 1 "task-master is NOT accessible in PATH"
fi
echo ""

# Test 7: Check task-master version
echo "Test 7: Checking task-master version..."
if command -v task-master &> /dev/null; then
    task-master --version
    print_test $? "task-master version check"
else
    print_test 1 "task-master not found"
fi
echo ""

# Test 8: Test idempotency - run installation again
echo "Test 8: Testing installation idempotency..."
make install-task-master
print_test $? "Repeated installation succeeds (idempotent)"
echo ""

# Test 9: Verify task-master still works after re-install
echo "Test 9: Verifying task-master still works after re-install..."
if command -v task-master &> /dev/null; then
    task-master --version > /dev/null 2>&1
    print_test $? "task-master works after re-install"
else
    print_test 1 "task-master not found after re-install"
fi
echo ""

# Test 10: Check for Windows-specific path handling
echo "Test 10: Testing path handling..."
echo "npm bin directory: $(npm config get prefix)/bin"
echo "Checking if binaries exist in expected location..."
if [ -f "$(npm config get prefix)/bin/task-master" ] || [ -f "$(npm config get prefix)/bin/task-master.cmd" ]; then
    print_test 0 "task-master binary found in npm bin directory"
else
    print_test 1 "task-master binary NOT found in expected location"
    echo "Contents of npm bin directory:"
    ls -la "$(npm config get prefix)/bin" || echo "Directory does not exist"
fi
echo ""

echo "=================================="
echo "Test Summary"
echo "=================================="
echo "All critical tests completed."
echo "Check output above for any failures."
