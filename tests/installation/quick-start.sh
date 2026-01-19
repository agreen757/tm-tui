#!/bin/bash
# Quick start script for Linux installation testing
# 
# This script checks Docker availability and provides guidance for running tests

set -e

echo "======================================"
echo "Linux Installation Test Quick Start"
echo "======================================"
echo ""

# Check if Docker is available
if command -v docker &> /dev/null; then
    echo "✓ Docker is installed"
    
    # Check if Docker daemon is running
    if docker ps &> /dev/null; then
        echo "✓ Docker daemon is running"
        echo ""
        echo "Ready to run tests!"
        echo ""
        echo "To run all tests:"
        echo "  ./run-linux-tests.sh"
        echo ""
        echo "To run individual tests:"
        echo "  cd ubuntu && docker build -f Dockerfile.apt -t tm-test-ubuntu-apt ."
        echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-ubuntu-apt"
        echo ""
        
        # Offer to run tests
        read -p "Run all tests now? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            ./run-linux-tests.sh
        fi
    else
        echo "✗ Docker daemon is not running"
        echo ""
        echo "Please start Docker:"
        echo "  macOS: open -a Docker"
        echo "  Linux: sudo systemctl start docker"
        echo ""
        echo "Then run this script again or execute:"
        echo "  ./run-linux-tests.sh"
    fi
else
    echo "✗ Docker is not installed"
    echo ""
    echo "Please install Docker:"
    echo "  macOS: brew install --cask docker"
    echo "  Linux: https://docs.docker.com/engine/install/"
    echo ""
    echo "After installing Docker, run this script again."
fi
