#!/bin/bash
# Setup script for Windows test environment
# Creates a Windows Server container or Git Bash simulation for testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common-helpers.sh"

log_test_start "Windows Test Environment Setup"

# Check Docker availability
check_docker || exit 1

# Check if Windows containers are available
log_info "Checking for Windows container support..."

# Try to detect Windows container capability
if docker info 2>/dev/null | grep -q "OSType: windows"; then
    log_test_pass "Windows containers are available"
    WINDOWS_CONTAINERS=true
else
    log_warning "Windows containers not available"
    log_info "Windows containers require:"
    log_info "  - Windows host with Docker Desktop"
    log_info "  - Switch to Windows containers in Docker Desktop settings"
    WINDOWS_CONTAINERS=false
fi

if [ "$WINDOWS_CONTAINERS" = true ]; then
    # Build Windows Server container
    log_info "Setting up Windows Server container with Git Bash"
    docker_build_image \
        "tm-test-windows:latest" \
        "$SCRIPT_DIR/Dockerfile.windows" \
        "$SCRIPT_DIR"
    
    log_test_pass "Windows test environment configured successfully"
    
    echo ""
    echo "========================================"
    echo "Windows Test Environment Ready"
    echo "========================================"
    echo ""
    echo "Available image:"
    echo "  - tm-test-windows:latest"
    echo ""
    echo "Run tests with:"
    echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-windows:latest"
    echo ""
    echo "Or use docker-compose with Windows profile:"
    echo "  cd tests/docker"
    echo "  docker-compose --profile windows up windows-gitbash"
    echo ""
else
    # Create Git Bash simulation using Alpine + bash
    log_info "Creating Git Bash simulation container (Linux-based)"
    
    # Create a minimal Dockerfile for Git Bash simulation
    cat > "$SCRIPT_DIR/Dockerfile.gitbash-sim" <<'EOF'
FROM alpine:3.18

# Install bash, git, and Node.js to simulate Git Bash environment
RUN apk add --no-cache \
    bash \
    git \
    make \
    curl \
    nodejs \
    npm

# Create a user similar to Windows environment
RUN adduser -D -s /bin/bash testuser

# Set up environment to simulate Git Bash
ENV PS1='\u@\h MINGW64 \w\n$ '
ENV SHELL=/bin/bash

WORKDIR /workspace

# Copy test script
COPY test-gitbash-sim.sh /test-gitbash-sim.sh
RUN chmod +x /test-gitbash-sim.sh

USER testuser

CMD ["/test-gitbash-sim.sh"]
EOF

    # Create test script for Git Bash simulation
    cat > "$SCRIPT_DIR/test-gitbash-sim.sh" <<'EOF'
#!/bin/bash
# Test script simulating Git Bash environment on Windows

source /workspace/tests/installation/common/test-helpers.sh

log_test_start "Git Bash Simulation Test"

log_info "Environment: $(uname -a)"
log_info "Shell: $SHELL"
log_info "User: $(whoami)"

# Test make installation
cd /workspace
log_info "Testing Task Master installation via make"

if make install-task-master 2>&1; then
    log_test_pass "make install-task-master succeeded"
else
    log_test_fail "make install-task-master failed"
fi

# Verify installation
if make check-task-master 2>&1; then
    log_test_pass "Task Master installation verified"
else
    log_test_fail "Task Master installation verification failed"
fi

print_test_summary
EOF

    chmod +x "$SCRIPT_DIR/test-gitbash-sim.sh"
    
    docker_build_image \
        "tm-test-gitbash-sim:latest" \
        "$SCRIPT_DIR/Dockerfile.gitbash-sim" \
        "$SCRIPT_DIR"
    
    log_test_pass "Git Bash simulation environment configured"
    
    echo ""
    echo "========================================"
    echo "Git Bash Simulation Environment Ready"
    echo "========================================"
    echo ""
    echo "Available image:"
    echo "  - tm-test-gitbash-sim:latest"
    echo ""
    echo "Run tests with:"
    echo "  docker run --rm -v \"\$(pwd)/../..:/workspace:ro\" tm-test-gitbash-sim:latest"
    echo ""
    echo "Note: This is a Linux-based simulation of Git Bash."
    echo "For true Windows testing, use a Windows host with Windows containers."
    echo ""
fi

print_test_summary
