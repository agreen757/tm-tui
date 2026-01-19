#!/bin/bash
# Common test helper functions for installation testing

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test result tracking
TESTS_PASSED=0
TESTS_FAILED=0

# Log functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_test_start() {
    echo ""
    echo "=================================="
    echo "TEST: $1"
    echo "=================================="
}

log_test_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_test_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

# Print test summary
print_test_summary() {
    echo ""
    echo "=================================="
    echo "TEST SUMMARY"
    echo "=================================="
    echo "Passed: $TESTS_PASSED"
    echo "Failed: $TESTS_FAILED"
    echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed.${NC}"
        return 1
    fi
}

# Verify command exists in PATH
check_command() {
    local cmd=$1
    if command -v "$cmd" >/dev/null 2>&1; then
        log_test_pass "Command '$cmd' found in PATH at: $(which "$cmd")"
        return 0
    else
        log_test_fail "Command '$cmd' not found in PATH"
        return 1
    fi
}

# Verify PATH contains specific directory
check_path_contains() {
    local dir=$1
    if echo "$PATH" | grep -q "$dir"; then
        log_test_pass "PATH contains: $dir"
        return 0
    else
        log_test_fail "PATH does not contain: $dir"
        log_info "Current PATH: $PATH"
        return 1
    fi
}

# Run a command and check exit code
run_and_check() {
    local description=$1
    shift
    local cmd="$@"
    
    log_info "Running: $cmd"
    if $cmd; then
        log_test_pass "$description"
        return 0
    else
        log_test_fail "$description (exit code: $?)"
        return 1
    fi
}

# Check npm configuration
check_npm_config() {
    log_info "NPM Configuration:"
    log_info "  npm prefix: $(npm config get prefix)"
    log_info "  npm bin location: $(npm bin -g 2>/dev/null || echo 'N/A')"
    
    local npm_prefix=$(npm config get prefix)
    local expected_bin="$npm_prefix/bin"
    
    if [ -d "$expected_bin" ]; then
        log_test_pass "NPM global bin directory exists: $expected_bin"
        
        # Check if task-master binary exists there
        if [ -f "$expected_bin/task-master" ]; then
            log_test_pass "task-master binary found in npm bin directory"
        else
            log_test_fail "task-master binary not found in npm bin directory"
        fi
    else
        log_test_fail "NPM global bin directory does not exist: $expected_bin"
    fi
}

# Export functions for use in other scripts
export -f log_info log_error log_warning log_test_start log_test_pass log_test_fail
export -f print_test_summary check_command check_path_contains run_and_check check_npm_config
