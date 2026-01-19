#!/bin/bash
# Setup script for Alpine Linux Docker test environment
# Configures Alpine containers with Node.js for Task Master testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common-helpers.sh"

log_test_start "Alpine Linux Docker Environment Setup"

# Check Docker availability
check_docker || exit 1

# Build Alpine with apk
log_info "Setting up Alpine Linux with apk-installed Node.js"
docker_build_image \
    "tm-test-alpine-apk:latest" \
    "$SCRIPT_DIR/../installation/alpine/Dockerfile.apk" \
    "$SCRIPT_DIR/../installation/alpine"

# Build Alpine with nvm
log_info "Setting up Alpine Linux with nvm-installed Node.js"
docker_build_image \
    "tm-test-alpine-nvm:latest" \
    "$SCRIPT_DIR/../installation/alpine/Dockerfile.nvm" \
    "$SCRIPT_DIR/../installation/alpine"

log_test_pass "Alpine Linux Docker environments configured successfully"

# Print usage information
echo ""
echo "========================================"
echo "Alpine Test Environments Ready"
echo "========================================"
echo ""
echo "Available images:"
echo "  - tm-test-alpine-apk:latest"
echo "  - tm-test-alpine-nvm:latest"
echo ""
echo "Run tests with:"
echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-alpine-apk:latest"
echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-alpine-nvm:latest"
echo ""
echo "Or use docker-compose:"
echo "  cd tests/docker"
echo "  docker-compose up alpine-apk alpine-nvm"
echo ""

print_test_summary
