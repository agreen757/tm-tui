#!/bin/bash
# Master test runner for Linux installation tests
# Tests Task Master installation across multiple Linux distributions and Node.js installation methods

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
RESULTS_DIR="$SCRIPT_DIR/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create results directory
mkdir -p "$RESULTS_DIR"

echo "=========================================="
echo "Task Master Linux Installation Test Suite"
echo "=========================================="
echo "Timestamp: $TIMESTAMP"
echo "Project Root: $PROJECT_ROOT"
echo ""

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test
run_test() {
    local test_name=$1
    local distro=$2
    local dockerfile=$3
    local log_file="$RESULTS_DIR/${test_name}_${TIMESTAMP}.log"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}Running: $test_name${NC}"
    echo "Log file: $log_file"
    
    # Build Docker image
    if docker build -f "$SCRIPT_DIR/$distro/$dockerfile" -t "tm-test-$test_name" "$SCRIPT_DIR/$distro" > "$log_file" 2>&1; then
        echo "  ✓ Docker image built successfully"
    else
        echo -e "  ${RED}✗ Failed to build Docker image${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
    
    # Run Docker container
    if docker run --rm -v "$PROJECT_ROOT:/workspace:ro" "tm-test-$test_name" >> "$log_file" 2>&1; then
        echo -e "  ${GREEN}✓ Test passed${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "  ${RED}✗ Test failed${NC}"
        echo -e "  ${YELLOW}Check log for details: $log_file${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Test 1: Ubuntu with apt
echo ""
echo "=========================================="
echo "Test 1: Ubuntu 22.04 + apt Node.js"
echo "=========================================="
run_test "ubuntu-apt" "ubuntu" "Dockerfile.apt" || true

# Test 2: Ubuntu with nvm
echo ""
echo "=========================================="
echo "Test 2: Ubuntu 22.04 + nvm Node.js"
echo "=========================================="
run_test "ubuntu-nvm" "ubuntu" "Dockerfile.nvm" || true

# Test 3: Alpine with apk
echo ""
echo "=========================================="
echo "Test 3: Alpine Linux + apk Node.js"
echo "=========================================="
run_test "alpine-apk" "alpine" "Dockerfile.apk" || true

# Test 4: Alpine with nvm
echo ""
echo "=========================================="
echo "Test 4: Alpine Linux + nvm Node.js"
echo "=========================================="
run_test "alpine-nvm" "alpine" "Dockerfile.nvm" || true

# Summary
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "Total tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo ""
echo "Results stored in: $RESULTS_DIR"

# Exit with appropriate status
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Some tests failed. Check logs for details.${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
