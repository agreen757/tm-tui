#!/usr/bin/env bash
# Individual test runner for specific edge cases
# Usage: ./run-single-test.sh <test_number>

set -euo pipefail

TEST_NUM="${1:-1}"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

cd "$PROJECT_ROOT"

case "$TEST_NUM" in
    1)
        echo "=== Test 1: Missing npm ==="
        docker build -f tests/edge-cases/dockerfiles/1-no-npm.Dockerfile -t tm-test-1 .
        docker run --rm -it -v "$(pwd):/workspace" tm-test-1 bash
        ;;
    2)
        echo "=== Test 2: Already installed ==="
        docker build -f tests/edge-cases/dockerfiles/2-already-installed.Dockerfile -t tm-test-2 .
        docker run --rm -it -v "$(pwd):/workspace" tm-test-2 bash
        ;;
    3)
        echo "=== Test 3: Misconfigured PATH ==="
        docker build -f tests/edge-cases/dockerfiles/3-misconfigured-path.Dockerfile -t tm-test-3 .
        docker run --rm -it -v "$(pwd):/workspace" tm-test-3 bash
        ;;
    4)
        echo "=== Test 4: Permission errors ==="
        docker build -f tests/edge-cases/dockerfiles/4-permission-errors.Dockerfile -t tm-test-4 .
        docker run --rm -it -v "$(pwd):/workspace" tm-test-4 bash
        ;;
    5)
        echo "=== Test 5: Old npm version ==="
        docker build -f tests/edge-cases/dockerfiles/5-old-npm.Dockerfile -t tm-test-5 .
        docker run --rm -it -v "$(pwd):/workspace" tm-test-5 bash
        ;;
    6)
        echo "=== Test 6: Limited disk space ==="
        docker build -f tests/edge-cases/dockerfiles/6-disk-space.Dockerfile -t tm-test-6 .
        docker run --rm -it -v "$(pwd):/workspace" tm-test-6 bash
        ;;
    7)
        echo "=== Test 7: Multiple global packages ==="
        docker build -f tests/edge-cases/dockerfiles/7-multiple-packages.Dockerfile -t tm-test-7 .
        docker run --rm -it -v "$(pwd):/workspace" tm-test-7 bash
        ;;
    *)
        echo "Unknown test number: $TEST_NUM"
        echo "Usage: $0 <1-7>"
        exit 1
        ;;
esac
