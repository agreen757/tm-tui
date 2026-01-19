#!/bin/bash
# Test Task Master installation on Ubuntu with Node.js via apt
set -e

echo "=========================================="
echo "Ubuntu + apt Node.js Installation Test"
echo "=========================================="

# Update package lists
echo "Updating package lists..."
apt-get update -qq

# Install Node.js and npm via apt
echo "Installing Node.js and npm via apt..."
apt-get install -y nodejs npm

# Verify Node.js installation
echo "Node.js version:"
node --version
echo "npm version:"
npm --version

# Navigate to project directory
cd /workspace

# Test make install-task-master
echo ""
echo "Testing: make install-task-master"
make install-task-master

# Verify installation
echo ""
echo "Verifying task-master installation..."
which task-master || echo "WARNING: task-master not in PATH"

# Check npm global bin directory
echo ""
echo "npm global bin directory:"
npm config get prefix

# Show PATH
echo ""
echo "Current PATH:"
echo "$PATH"

# Try to run task-master
echo ""
echo "Testing task-master command..."
if command -v task-master &> /dev/null; then
    task-master --version || echo "task-master exists but --version failed"
    
    # Check binary permissions
    echo ""
    echo "Binary location and permissions:"
    ls -l "$(which task-master)"
else
    echo "ERROR: task-master command not found in PATH"
    echo "Checking npm global directory..."
    npm list -g --depth=0 | grep task-master || echo "task-master not in global packages"
    
    # Try to find it manually
    echo "Searching for task-master binary..."
    find /usr -name task-master 2>/dev/null || echo "Not found in /usr"
fi

# Test idempotency - run installation again
echo ""
echo "=========================================="
echo "Testing idempotency (second installation)"
echo "=========================================="
make install-task-master

echo ""
echo "Second installation complete - checking status..."
which task-master || echo "WARNING: task-master still not in PATH after second install"

# Test make check-task-master
echo ""
echo "Testing: make check-task-master"
make check-task-master || echo "check-task-master target failed"

echo ""
echo "=========================================="
echo "Ubuntu + apt test completed"
echo "=========================================="
