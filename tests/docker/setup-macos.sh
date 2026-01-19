#!/bin/bash
# Setup script for macOS Docker test environment
# Since Docker on macOS runs Linux containers, this script helps test macOS-specific scenarios

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common-helpers.sh"

log_test_start "macOS Docker Environment Setup"

# Check if running on macOS
if [[ "$(uname)" != "Darwin" ]]; then
    log_warning "This script is designed for macOS"
    log_info "For testing macOS installation, run directly on macOS host:"
    log_info "  cd tests/installation"
    log_info "  ./test-local-macos.sh"
    exit 0
fi

# Check Docker Desktop availability
check_docker || exit 1

# Verify Docker Desktop is running (macOS-specific)
if pgrep -f "Docker Desktop" >/dev/null; then
    log_test_pass "Docker Desktop is running"
else
    log_warning "Docker Desktop may not be running"
    log_info "Start Docker Desktop: open -a Docker"
fi

# Check Docker VM resources
log_info "Docker Desktop Configuration:"
docker info | grep -E "CPUs|Total Memory|Operating System" || true

echo ""
echo "========================================"
echo "macOS Testing Options"
echo "========================================"
echo ""
echo "Option 1: Test on macOS host directly"
echo "  cd ../installation"
echo "  ./test-local-macos.sh"
echo ""
echo "Option 2: Test Linux containers on macOS"
echo "  cd tests/docker"
echo "  ./setup-ubuntu.sh"
echo "  ./setup-alpine.sh"
echo ""
echo "Option 3: Use docker-compose for all tests"
echo "  cd tests/docker"
echo "  docker-compose up"
echo ""
echo "Note: Docker Desktop on macOS runs Linux containers."
echo "For native macOS testing, use Option 1."
echo ""

# Optionally set up Linux test containers
read -p "Set up Linux containers for testing on macOS? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Setting up Linux test containers..."
    "$SCRIPT_DIR/setup-ubuntu.sh"
    "$SCRIPT_DIR/setup-alpine.sh"
fi

print_test_summary
