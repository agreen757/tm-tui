#!/bin/bash
# Test script for nvm-style installation (per-user Node.js)

set -e

# Source helper functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../common/test-helpers.sh"

log_test_start "NVM-style Installation Test"

# Test 1: Verify Node.js and npm are installed via nvm
log_info "Step 1: Verify Node.js and npm are available"
check_command node
check_command npm

# Test 2: Check npm configuration (should point to nvm directory)
log_info "Step 2: Check npm configuration"
check_npm_config

NPM_PREFIX=$(npm config get prefix)
log_info "NPM prefix: $NPM_PREFIX"

# Test 3: Verify npm prefix is in user's home directory (nvm-style)
if [[ "$NPM_PREFIX" == *"$HOME"* ]] || [[ "$NPM_PREFIX" == *".nvm"* ]]; then
    log_test_pass "NPM prefix is in user directory (nvm-style installation)"
else
    log_warning "NPM prefix is not in user directory: $NPM_PREFIX"
fi

# Test 4: Run make install-task-master (first time)
log_test_start "First Installation Run"
log_info "Step 4: Install Task Master CLI via make target"
run_and_check "make install-task-master" make install-task-master

# Test 5: Verify task-master is in PATH
log_info "Step 5: Verify task-master is accessible"
check_command task-master

# Test 6: Verify PATH contains nvm bin directory
log_info "Step 6: Verify PATH configuration for nvm"
if echo "$PATH" | grep -q ".nvm"; then
    log_test_pass "PATH contains nvm directory"
else
    log_test_fail "PATH does not contain nvm directory"
fi

# Test 7: Test task-master functionality
log_test_start "Task Master Functionality Tests"
if [ -d ".taskmaster" ]; then
    log_info "Step 7a: Test task-master list command"
    run_and_check "task-master list" task-master list
    
    log_info "Step 7b: Test task-master show command"
    run_and_check "task-master show 1" task-master show 1 2>/dev/null || log_warning "No task 1 available (expected in test environment)"
else
    log_warning "No .taskmaster directory found, skipping task-master functionality tests"
fi

# Test 8: Test idempotency (run installation again)
log_test_start "Idempotency Test"
log_info "Step 8: Run installation again (should succeed without errors)"
run_and_check "make install-task-master (second run)" make install-task-master

# Test 9: Verify task-master still works after second installation
log_info "Step 9: Verify task-master still accessible after second installation"
check_command task-master

# Test 10: Verify binary location is in nvm directory
log_info "Step 10: Verify task-master binary location"
TASK_MASTER_PATH=$(which task-master)
if [[ "$TASK_MASTER_PATH" == *".nvm"* ]]; then
    log_test_pass "task-master is installed in nvm directory: $TASK_MASTER_PATH"
else
    log_warning "task-master is not in nvm directory (found at: $TASK_MASTER_PATH)"
fi

# Print summary
print_test_summary
