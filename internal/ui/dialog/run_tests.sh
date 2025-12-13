#!/bin/sh

# Script to run dialog component tests and validation

set -e  # Exit on any error

echo "===== Running Dialog Component Tests ====="
echo

# Run all the tests with verbose output
go test -v .

echo
echo "===== Testing Complete ====="

# Check if tests passed
if [ $? -eq 0 ]; then
    echo "All tests passed!"
    echo
    echo "To run the dialog demo screen:"
    echo "go run internal/ui/dialog/demo/main/main.go"
else
    echo "Some tests failed. Please check the output above."
    exit 1
fi