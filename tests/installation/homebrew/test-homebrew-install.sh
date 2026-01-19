#!/bin/bash
# Test script for Homebrew-style installation (system-wide Node.js)

set -e

# Source helper functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../common/test-helpers.sh"

log_test_start "Homebrew-style Installation Test"

# Test 1: Verify Node.js and npm are installed
log_info "Step 1: Verify Node.js and npm are available"
check_command node
check_command npm

# Test 2: Check npm configuration
log_info "Step 2: Check npm configuration"
check_npm_config

# Test 3: Run make install-task-master (first time)
log_test_start "First Installation Run"
log_info "Step 3: Install Task Master CLI via make target"
run_and_check "make install-task-master" make install-task-master

# Test 4: Verify task-master is in PATH
log_info "Step 4: Verify task-master is accessible"
check_command task-master

# Test 5: Test task-master functionality
log_test_start "Task Master Functionality Tests"
if [ -d ".taskmaster" ]; then
    log_info "Step 5a: Test task-master list command"
    run_and_check "task-master list" task-master list
    
    log_info "Step 5b: Test task-master show command"
    run_and_check "task-master show 1" task-master show 1 2>/dev/null || log_warning "No task 1 available (expected in test environment)"
else
    log_warning "No .taskmaster directory found, skipping task-master functionality tests"
fi

# Test 6: Test idempotency (run installation again)
log_test_start "Idempotency Test"
log_info "Step 6: Run installation again (should succeed without errors)"
run_and_check "make install-task-master (second run)" make install-task-master

# Test 7: Verify task-master still works after second installation
log_info "Step 7: Verify task-master still accessible after second installation"
check_command task-master

# Test 8: Check PATH contains npm bin directory
log_info "Step 8: Verify PATH configuration"
NPM_PREFIX=$(npm config get prefix)
check_path_contains "$NPM_PREFIX/bin"

# Test 9: Verify binary location matches npm prefix
log_info "Step 9: Verify task-master binary location"
TASK_MASTER_PATH=$(which task-master)
if [[ "$TASK_MASTER_PATH" == "$NPM_PREFIX"* ]]; then
    log_test_pass "task-master is installed in npm prefix directory"
else
    log_test_fail "task-master is not in npm prefix directory (found at: $TASK_MASTER_PATH)"
fi

# Print summary
print_test_summary
