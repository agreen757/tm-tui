#!/bin/bash
# Master test script for installation testing on macOS (via Docker)
# 
# This script tests the installation workflow with two Node.js installation methods:
# 1. Homebrew-style (system-wide Node.js)
# 2. nvm-style (per-user Node.js via version manager)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=================================="
echo "Installation Testing Suite"
echo "=================================="
echo ""
echo "Project root: $PROJECT_ROOT"
echo ""

# Track overall results
SCENARIOS_PASSED=0
SCENARIOS_FAILED=0

# Test Scenario 1: Homebrew-style installation
echo ""
echo "######################################"
echo "# Scenario 1: Homebrew-style Install"
echo "######################################"
echo ""

cd "$PROJECT_ROOT"

echo "Building Docker image for Homebrew test..."
docker build -t tm-tui-test-homebrew -f tests/installation/homebrew/Dockerfile .

echo ""
echo "Running Homebrew installation test..."
if docker run --rm \
    -v "$PROJECT_ROOT:/workspace" \
    -w /workspace \
    tm-tui-test-homebrew \
    bash tests/installation/homebrew/test-homebrew-install.sh; then
    echo ""
    echo "✓ Homebrew-style installation test PASSED"
    SCENARIOS_PASSED=$((SCENARIOS_PASSED + 1))
else
    echo ""
    echo "✗ Homebrew-style installation test FAILED"
    SCENARIOS_FAILED=$((SCENARIOS_FAILED + 1))
fi

# Test Scenario 2: nvm-style installation
echo ""
echo "######################################"
echo "# Scenario 2: NVM-style Install"
echo "######################################"
echo ""

cd "$PROJECT_ROOT"

echo "Building Docker image for nvm test..."
docker build -t tm-tui-test-nvm -f tests/installation/nvm/Dockerfile .

echo ""
echo "Running nvm installation test..."
if docker run --rm \
    -v "$PROJECT_ROOT:/workspace" \
    -w /workspace \
    tm-tui-test-nvm \
    bash tests/installation/nvm/test-nvm-install.sh; then
    echo ""
    echo "✓ NVM-style installation test PASSED"
    SCENARIOS_PASSED=$((SCENARIOS_PASSED + 1))
else
    echo ""
    echo "✗ NVM-style installation test FAILED"
    SCENARIOS_FAILED=$((SCENARIOS_FAILED + 1))
fi

# Print final summary
echo ""
echo "========================================"
echo "INSTALLATION TEST SUITE SUMMARY"
echo "========================================"
echo "Scenarios Passed: $SCENARIOS_PASSED"
echo "Scenarios Failed: $SCENARIOS_FAILED"
echo "Total Scenarios:  $((SCENARIOS_PASSED + SCENARIOS_FAILED))"
echo ""

if [ $SCENARIOS_FAILED -eq 0 ]; then
    echo "✓ All installation scenarios passed!"
    exit 0
else
    echo "✗ Some installation scenarios failed."
    exit 1
fi
