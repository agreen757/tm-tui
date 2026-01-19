#!/bin/bash
# Setup script for Ubuntu Docker test environment
# Configures Ubuntu containers with Node.js for Task Master testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common-helpers.sh"

log_test_start "Ubuntu Docker Environment Setup"

# Check Docker availability
check_docker || exit 1

# Build Ubuntu with apt
log_info "Setting up Ubuntu 22.04 with apt-installed Node.js"
docker_build_image \
    "tm-test-ubuntu-apt:latest" \
    "$SCRIPT_DIR/../installation/ubuntu/Dockerfile.apt" \
    "$SCRIPT_DIR/../installation/ubuntu"

# Build Ubuntu with nvm
log_info "Setting up Ubuntu 22.04 with nvm-installed Node.js"
docker_build_image \
    "tm-test-ubuntu-nvm:latest" \
    "$SCRIPT_DIR/../installation/ubuntu/Dockerfile.nvm" \
    "$SCRIPT_DIR/../installation/ubuntu"

log_test_pass "Ubuntu Docker environments configured successfully"

# Print usage information
echo ""
echo "========================================"
echo "Ubuntu Test Environments Ready"
echo "========================================"
echo ""
echo "Available images:"
echo "  - tm-test-ubuntu-apt:latest"
echo "  - tm-test-ubuntu-nvm:latest"
echo ""
echo "Run tests with:"
echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-ubuntu-apt:latest"
echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-ubuntu-nvm:latest"
echo ""
echo "Or use docker-compose:"
echo "  cd tests/docker"
echo "  docker-compose up ubuntu-apt ubuntu-nvm"
echo ""

print_test_summary
