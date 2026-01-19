#!/bin/bash
# Master setup script for all Docker test environments
# Orchestrates the creation and configuration of all platform test containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common-helpers.sh"

echo "=========================================="
echo "Task Master Docker Test Environment Setup"
echo "=========================================="
echo ""
echo "This script will set up Docker test environments for:"
echo "  - Ubuntu 22.04 (apt and nvm)"
echo "  - Alpine Linux (apk and nvm)"
echo "  - macOS (via Docker Desktop)"
echo "  - Windows/Git Bash simulation"
echo ""

# Check Docker prerequisites
check_docker || exit 1
check_docker_compose

echo ""
read -p "Continue with setup? (Y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo "Setup cancelled."
    exit 0
fi

# Track setup results
SETUPS_PASSED=0
SETUPS_FAILED=0

# Setup Ubuntu
echo ""
echo "=========================================="
echo "Setting up Ubuntu environments..."
echo "=========================================="
if "$SCRIPT_DIR/setup-ubuntu.sh"; then
    SETUPS_PASSED=$((SETUPS_PASSED + 1))
else
    SETUPS_FAILED=$((SETUPS_FAILED + 1))
    log_error "Ubuntu setup failed"
fi

# Setup Alpine
echo ""
echo "=========================================="
echo "Setting up Alpine Linux environments..."
echo "=========================================="
if "$SCRIPT_DIR/setup-alpine.sh"; then
    SETUPS_PASSED=$((SETUPS_PASSED + 1))
else
    SETUPS_FAILED=$((SETUPS_FAILED + 1))
    log_error "Alpine setup failed"
fi

# Setup macOS (if on macOS)
if [[ "$(uname)" == "Darwin" ]]; then
    echo ""
    echo "=========================================="
    echo "Configuring macOS test environment..."
    echo "=========================================="
    if "$SCRIPT_DIR/setup-macos.sh"; then
        SETUPS_PASSED=$((SETUPS_PASSED + 1))
    else
        SETUPS_FAILED=$((SETUPS_FAILED + 1))
        log_error "macOS setup failed"
    fi
fi

# Setup Windows/Git Bash simulation
echo ""
echo "=========================================="
echo "Setting up Windows test environment..."
echo "=========================================="
if "$SCRIPT_DIR/setup-windows.sh"; then
    SETUPS_PASSED=$((SETUPS_PASSED + 1))
else
    SETUPS_FAILED=$((SETUPS_FAILED + 1))
    log_error "Windows setup failed"
fi

# Summary
echo ""
echo "=========================================="
echo "Setup Summary"
echo "=========================================="
echo "Environments set up successfully: $SETUPS_PASSED"
echo "Environments failed: $SETUPS_FAILED"
echo ""

# List available images
echo "=========================================="
echo "Available Docker Images"
echo "=========================================="
docker images | grep "tm-test-" || echo "No test images found"
echo ""

# Print usage instructions
echo "=========================================="
echo "Usage Instructions"
echo "=========================================="
echo ""
echo "Run all tests with docker-compose:"
echo "  cd $SCRIPT_DIR"
echo "  docker-compose up"
echo ""
echo "Run individual platform tests:"
echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-ubuntu-apt:latest"
echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-alpine-apk:latest"
echo ""
echo "Run specific services:"
echo "  docker-compose up ubuntu-apt ubuntu-nvm"
echo "  docker-compose up alpine-apk"
echo ""
echo "Clean up all test containers and images:"
echo "  docker-compose down"
echo "  docker rmi \$(docker images -q 'tm-test-*')"
echo ""

if [ $SETUPS_FAILED -eq 0 ]; then
    echo "✓ All environments set up successfully!"
    exit 0
else
    echo "✗ Some environments failed to set up. Check logs above."
    exit 1
fi
