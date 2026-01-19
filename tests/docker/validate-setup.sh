#!/bin/bash
# Quick validation script to test Docker setup scripts without building images
# Verifies syntax, structure, and basic functionality

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=========================================="
echo "Docker Setup Scripts Validation"
echo "=========================================="
echo ""

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
run_test() {
    local test_name=$1
    shift
    local test_cmd="$@"
    
    echo -n "Testing $test_name... "
    if eval "$test_cmd" >/dev/null 2>&1; then
        echo "✓ PASS"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo "✗ FAIL"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Check file existence
run_test "docker-compose.yml exists" "[ -f '$SCRIPT_DIR/docker-compose.yml' ]"
run_test "common-helpers.sh exists" "[ -f '$SCRIPT_DIR/common-helpers.sh' ]"
run_test "setup-all.sh exists" "[ -f '$SCRIPT_DIR/setup-all.sh' ]"
run_test "setup-ubuntu.sh exists" "[ -f '$SCRIPT_DIR/setup-ubuntu.sh' ]"
run_test "setup-alpine.sh exists" "[ -f '$SCRIPT_DIR/setup-alpine.sh' ]"
run_test "setup-macos.sh exists" "[ -f '$SCRIPT_DIR/setup-macos.sh' ]"
run_test "setup-windows.sh exists" "[ -f '$SCRIPT_DIR/setup-windows.sh' ]"
run_test "Dockerfile.windows exists" "[ -f '$SCRIPT_DIR/Dockerfile.windows' ]"
run_test "test-windows.bat exists" "[ -f '$SCRIPT_DIR/test-windows.bat' ]"
run_test "README.md exists" "[ -f '$SCRIPT_DIR/README.md' ]"

# Check script executability
run_test "setup-all.sh is executable" "[ -x '$SCRIPT_DIR/setup-all.sh' ]"
run_test "setup-ubuntu.sh is executable" "[ -x '$SCRIPT_DIR/setup-ubuntu.sh' ]"
run_test "setup-alpine.sh is executable" "[ -x '$SCRIPT_DIR/setup-alpine.sh' ]"
run_test "setup-macos.sh is executable" "[ -x '$SCRIPT_DIR/setup-macos.sh' ]"
run_test "setup-windows.sh is executable" "[ -x '$SCRIPT_DIR/setup-windows.sh' ]"
run_test "common-helpers.sh is executable" "[ -x '$SCRIPT_DIR/common-helpers.sh' ]"

# Check bash syntax
run_test "setup-all.sh syntax" "bash -n '$SCRIPT_DIR/setup-all.sh'"
run_test "setup-ubuntu.sh syntax" "bash -n '$SCRIPT_DIR/setup-ubuntu.sh'"
run_test "setup-alpine.sh syntax" "bash -n '$SCRIPT_DIR/setup-alpine.sh'"
run_test "setup-macos.sh syntax" "bash -n '$SCRIPT_DIR/setup-macos.sh'"
run_test "setup-windows.sh syntax" "bash -n '$SCRIPT_DIR/setup-windows.sh'"
run_test "common-helpers.sh syntax" "bash -n '$SCRIPT_DIR/common-helpers.sh'"

# Check docker-compose syntax (if Docker is available)
if command -v docker >/dev/null 2>&1; then
    if docker compose version >/dev/null 2>&1; then
        run_test "docker-compose.yml syntax" "cd '$SCRIPT_DIR' && docker compose config >/dev/null 2>&1"
    elif command -v docker-compose >/dev/null 2>&1; then
        run_test "docker-compose.yml syntax" "cd '$SCRIPT_DIR' && docker-compose config >/dev/null 2>&1"
    else
        echo "Skipping docker-compose validation (not available)"
    fi
else
    echo "Skipping Docker tests (Docker not available)"
fi

# Test helper functions
run_test "common-helpers.sh sources" "bash -c 'source $SCRIPT_DIR/common-helpers.sh 2>/dev/null'"

# Check referenced paths exist
run_test "../installation/ubuntu exists" "[ -d '$SCRIPT_DIR/../installation/ubuntu' ]"
run_test "../installation/alpine exists" "[ -d '$SCRIPT_DIR/../installation/alpine' ]"
run_test "../installation/common/test-helpers.sh exists" "[ -f '$SCRIPT_DIR/../installation/common/test-helpers.sh' ]"

echo ""
echo "=========================================="
echo "Validation Summary"
echo "=========================================="
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"
echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo "✓ All validation tests passed!"
    echo ""
    echo "Setup scripts are ready to use."
    echo "Run './setup-all.sh' to build Docker images."
    exit 0
else
    echo "✗ Some validation tests failed."
    exit 1
fi
