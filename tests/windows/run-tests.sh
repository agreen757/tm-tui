#!/bin/bash
# Automated test runner for Windows installation testing

set -e

echo "========================================"
echo "Task Master TUI - Windows Install Tests"
echo "========================================"
echo ""

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"

echo "Project root: $PROJECT_ROOT"
echo ""

# Step 1: Build Docker image
echo "${YELLOW}[1/3] Building Docker image...${NC}"
docker build -f tests/windows/Dockerfile.windows-test -t tm-tui-windows-test . || {
    echo "${RED}❌ Docker build failed${NC}"
    exit 1
}
echo "${GREEN}✅ Docker image built successfully${NC}"
echo ""

# Step 2: Run tests in container
echo "${YELLOW}[2/3] Running installation tests...${NC}"
docker run --rm \
  -v "$PROJECT_ROOT":/home/testuser/workspace \
  tm-tui-windows-test \
  /home/testuser/workspace/tests/windows/test-installation.sh || {
    echo "${RED}❌ Tests failed${NC}"
    exit 1
}
echo ""

# Step 3: Summary
echo "${GREEN}✅ All tests completed successfully!${NC}"
echo ""
echo "Test results summary:"
echo "  - Environment setup: ✅"
echo "  - PATH configuration: ✅"
echo "  - Installation workflow: ✅"
echo "  - Idempotency: ✅"
echo "  - Binary accessibility: ✅"
echo ""
echo "See above for detailed test output."
