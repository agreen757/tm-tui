#!/bin/bash
# Test script for Crush config initialization behavior
# This script verifies that the application and Makefile targets work correctly
# across different scenarios.

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "========================================"
echo "Testing Crush Config Initialization"
echo "========================================"
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

# Test 1: Build the application
echo "Test 1: Building application..."
cd "$PROJECT_ROOT"
make build > /dev/null 2>&1 || fail "Build failed"
pass "Application builds successfully"
echo

# Test 2: check-project-setup with existing config
echo "Test 2: Testing check-project-setup with existing config..."
if [ -f .crush.json ]; then
    make check-project-setup > /dev/null 2>&1 || fail "check-project-setup failed with existing config"
    pass "check-project-setup succeeds with existing config"
else
    echo "  (Skipping: no .crush.json in project root)"
fi
echo

# Test 3: check-project-setup without config
echo "Test 3: Testing check-project-setup without config..."
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
cp "$PROJECT_ROOT/Makefile" .
if make check-project-setup > /dev/null 2>&1; then
    fail "check-project-setup should fail without config"
else
    pass "check-project-setup correctly fails without config"
fi
cd "$PROJECT_ROOT"
rm -rf "$TEST_DIR"
echo

# Test 4: init-crush-config creates config
echo "Test 4: Testing init-crush-config creates config..."
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
cp -r "$PROJECT_ROOT"/{Makefile,go.mod,go.sum,cmd,internal} . 2>/dev/null
make init-crush-config > /dev/null 2>&1 || fail "init-crush-config failed"
if [ -f .crush.json ]; then
    pass "init-crush-config creates .crush.json"
else
    fail ".crush.json not created"
fi
cd "$PROJECT_ROOT"
rm -rf "$TEST_DIR"
echo

# Test 5: init-crush-config preserves existing config
echo "Test 5: Testing init-crush-config preserves existing config..."
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
cp -r "$PROJECT_ROOT"/{Makefile,go.mod,go.sum,cmd,internal} . 2>/dev/null
echo '{"model": "test", "custom": "data"}' > .crush.json
BEFORE_CONTENT=$(cat .crush.json)
make init-crush-config > /dev/null 2>&1 || fail "init-crush-config failed with existing config"
AFTER_CONTENT=$(cat .crush.json)
if [ "$BEFORE_CONTENT" = "$AFTER_CONTENT" ]; then
    pass "init-crush-config preserves existing config"
else
    fail "init-crush-config modified existing config"
fi
cd "$PROJECT_ROOT"
rm -rf "$TEST_DIR"
echo

# Test 6: Application startup creates config
echo "Test 6: Testing application startup initialization..."
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
mkdir -p .taskmaster  # Project root marker
"$PROJECT_ROOT/bin/tm-tui" --help > /dev/null 2>&1 || true
if [ -f .crush.json ]; then
    pass "Application startup creates .crush.json"
else
    fail "Application startup did not create .crush.json"
fi
cd "$PROJECT_ROOT"
rm -rf "$TEST_DIR"
echo

# Test 7: Application startup preserves existing config
echo "Test 7: Testing application startup preserves existing config..."
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
mkdir -p .taskmaster
echo '{"model": "startup-test", "preserved": true}' > .crush.json
BEFORE_CONTENT=$(cat .crush.json)
"$PROJECT_ROOT/bin/tm-tui" --help > /dev/null 2>&1 || true
AFTER_CONTENT=$(cat .crush.json)
if [ "$BEFORE_CONTENT" = "$AFTER_CONTENT" ]; then
    pass "Application startup preserves existing config"
else
    fail "Application startup modified existing config"
fi
cd "$PROJECT_ROOT"
rm -rf "$TEST_DIR"
echo

# Test 8: Run Go tests
echo "Test 8: Running Go unit tests..."
cd "$PROJECT_ROOT"
if go test ./internal/config/... > /dev/null 2>&1; then
    pass "Config package unit tests pass"
else
    fail "Config package unit tests failed"
fi
echo

# Test 9: Run integration tests
echo "Test 9: Running integration tests..."
cd "$PROJECT_ROOT"
if go test ./cmd/tm-tui/... > /dev/null 2>&1; then
    pass "Main package integration tests pass"
else
    fail "Main package integration tests failed"
fi
echo

echo "========================================"
echo "All tests passed!"
echo "========================================"
