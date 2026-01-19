#!/bin/bash
# check-coverage.sh - Verify test coverage meets minimum threshold
# Usage: ./scripts/check-coverage.sh <coverage-file> <threshold>

set -e

COVERAGE_FILE=${1:-coverage.out}
THRESHOLD=${2:-80}

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "Error: Coverage file '$COVERAGE_FILE' not found"
    exit 1
fi

# Extract total coverage percentage
COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')

# Compare coverage with threshold
if [ -z "$COVERAGE" ]; then
    echo "Error: Could not extract coverage percentage"
    exit 1
fi

echo "Current coverage: ${COVERAGE}%"
echo "Required threshold: ${THRESHOLD}%"

# Use bc for floating point comparison
if (( $(echo "$COVERAGE >= $THRESHOLD" | bc -l) )); then
    echo "✓ Coverage meets threshold"
    exit 0
else
    echo "✗ Coverage below threshold (${COVERAGE}% < ${THRESHOLD}%)"
    exit 1
fi
