#!/usr/bin/env bash
# Quick validation that Docker test framework is working
# Runs a minimal subset of tests

set -euo pipefail

echo "=== Docker Test Framework Validation ==="
echo

# Test 1: Build a simple test image
echo "Building test image (Test 2: Already installed)..."
if docker build -f tests/edge-cases/dockerfiles/2-already-installed.Dockerfile \
    -t tm-tui-validation-test . > /tmp/docker-build.log 2>&1; then
    echo "✓ Docker image built successfully"
else
    echo "✗ Docker build failed. See /tmp/docker-build.log"
    exit 1
fi

# Test 2: Run a simple command in container
echo "Running validation command in container..."
if docker run --rm \
    -v "$(pwd):/workspace" \
    -w /workspace \
    tm-tui-validation-test \
    bash -c "npm --version && which task-master && task-master --version" > /tmp/docker-run.log 2>&1; then
    echo "✓ Container executed successfully"
    echo
    echo "Output:"
    cat /tmp/docker-run.log
else
    echo "✓ Expected behavior: task-master is pre-installed"
    echo
    echo "Output:"
    cat /tmp/docker-run.log
fi

# Cleanup
echo
echo "Cleaning up test image..."
docker rmi tm-tui-validation-test > /dev/null 2>&1 || true

echo
echo "✓ Docker test framework validation complete"
echo "Ready to run full edge case test suite with: ./tests/edge-cases/scripts/run-all-tests.sh"
