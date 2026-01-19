#!/bin/bash
# Common helper functions for Docker-based testing
# Extends the basic test-helpers.sh with Docker-specific operations

# Source the basic test helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/../installation/common/test-helpers.sh" ]; then
    source "$SCRIPT_DIR/../installation/common/test-helpers.sh"
fi

# Docker-specific color codes (if not already defined)
BLUE="${BLUE:-\033[0;34m}"
CYAN="${CYAN:-\033[0;36m}"

# Check if Docker is available and running
check_docker() {
    log_test_start "Docker availability check"
    
    if ! command -v docker >/dev/null 2>&1; then
        log_test_fail "Docker is not installed"
        log_info "Please install Docker:"
        log_info "  macOS: brew install --cask docker"
        log_info "  Ubuntu: sudo apt install docker.io"
        log_info "  Docs: https://docs.docker.com/get-docker/"
        return 1
    fi
    
    log_test_pass "Docker CLI found"
    
    if ! docker ps >/dev/null 2>&1; then
        log_test_fail "Docker daemon is not running"
        log_info "Please start Docker:"
        log_info "  macOS: open -a Docker"
        log_info "  Linux: sudo systemctl start docker"
        return 1
    fi
    
    log_test_pass "Docker daemon is running"
    
    local docker_version=$(docker --version)
    log_info "Docker version: $docker_version"
    
    return 0
}

# Check if docker-compose is available
check_docker_compose() {
    if command -v docker-compose >/dev/null 2>&1; then
        log_test_pass "docker-compose found"
        log_info "Version: $(docker-compose --version)"
        return 0
    elif docker compose version >/dev/null 2>&1; then
        log_test_pass "docker compose (plugin) found"
        log_info "Version: $(docker compose version)"
        return 0
    else
        log_test_fail "docker-compose or docker compose not found"
        return 1
    fi
}

# Build a Docker image with error handling
docker_build_image() {
    local image_name=$1
    local dockerfile=$2
    local context=${3:-.}
    
    log_info "Building Docker image: $image_name"
    log_info "  Dockerfile: $dockerfile"
    log_info "  Context: $context"
    
    if docker build -f "$dockerfile" -t "$image_name" "$context"; then
        log_test_pass "Docker image built: $image_name"
        return 0
    else
        log_test_fail "Failed to build Docker image: $image_name"
        return 1
    fi
}

# Run a Docker container with standard options
docker_run_test() {
    local image_name=$1
    local test_name=${2:-$image_name}
    shift 2
    local extra_args="$@"
    
    log_info "Running Docker test: $test_name"
    log_info "  Image: $image_name"
    
    if docker run --rm $extra_args "$image_name"; then
        log_test_pass "Docker test passed: $test_name"
        return 0
    else
        log_test_fail "Docker test failed: $test_name"
        return 1
    fi
}

# Check if a Docker container is running
check_container_running() {
    local container_name=$1
    
    if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
        log_test_pass "Container is running: $container_name"
        return 0
    else
        log_test_fail "Container is not running: $container_name"
        return 1
    fi
}

# Check container health status
check_container_health() {
    local container_name=$1
    
    local health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null)
    
    if [ -z "$health_status" ]; then
        log_warning "Container has no health check: $container_name"
        return 0
    fi
    
    if [ "$health_status" = "healthy" ]; then
        log_test_pass "Container is healthy: $container_name"
        return 0
    else
        log_test_fail "Container health: $health_status ($container_name)"
        return 1
    fi
}

# Execute command in running container
docker_exec_test() {
    local container_name=$1
    local test_description=$2
    shift 2
    local cmd="$@"
    
    log_info "Executing in container $container_name: $cmd"
    
    if docker exec "$container_name" $cmd; then
        log_test_pass "$test_description"
        return 0
    else
        log_test_fail "$test_description"
        return 1
    fi
}

# Clean up Docker resources
docker_cleanup() {
    local pattern=${1:-tm-test-}
    
    log_info "Cleaning up Docker resources matching: $pattern"
    
    # Stop and remove containers
    local containers=$(docker ps -a --format '{{.Names}}' | grep "^$pattern" || true)
    if [ -n "$containers" ]; then
        echo "$containers" | xargs docker rm -f >/dev/null 2>&1 || true
        log_info "Removed containers matching: $pattern"
    fi
    
    # Remove images
    local images=$(docker images --format '{{.Repository}}:{{.Tag}}' | grep "^$pattern" || true)
    if [ -n "$images" ]; then
        echo "$images" | xargs docker rmi -f >/dev/null 2>&1 || true
        log_info "Removed images matching: $pattern"
    fi
    
    log_test_pass "Docker cleanup completed"
}

# Wait for container to be healthy
wait_for_container() {
    local container_name=$1
    local timeout=${2:-30}
    local interval=2
    local elapsed=0
    
    log_info "Waiting for container to be ready: $container_name"
    
    while [ $elapsed -lt $timeout ]; do
        if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
            # Check if container has health check
            local health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null)
            
            if [ -z "$health_status" ]; then
                # No health check, just verify it's running
                log_test_pass "Container is running: $container_name"
                return 0
            elif [ "$health_status" = "healthy" ]; then
                log_test_pass "Container is healthy: $container_name"
                return 0
            fi
        fi
        
        sleep $interval
        elapsed=$((elapsed + interval))
        echo -n "."
    done
    
    echo ""
    log_test_fail "Timeout waiting for container: $container_name"
    return 1
}

# Get container logs
get_container_logs() {
    local container_name=$1
    local lines=${2:-50}
    
    log_info "Container logs (last $lines lines): $container_name"
    docker logs --tail $lines "$container_name" 2>&1 || log_error "Failed to get logs"
}

# Run docker-compose up with proper error handling
compose_up() {
    local compose_file=${1:-docker-compose.yml}
    local services="${2:-}"
    
    log_info "Starting services with docker-compose"
    log_info "  Compose file: $compose_file"
    
    if command -v docker-compose >/dev/null 2>&1; then
        docker-compose -f "$compose_file" up -d $services
    else
        docker compose -f "$compose_file" up -d $services
    fi
    
    if [ $? -eq 0 ]; then
        log_test_pass "Services started successfully"
        return 0
    else
        log_test_fail "Failed to start services"
        return 1
    fi
}

# Run docker-compose down with cleanup
compose_down() {
    local compose_file=${1:-docker-compose.yml}
    
    log_info "Stopping services with docker-compose"
    
    if command -v docker-compose >/dev/null 2>&1; then
        docker-compose -f "$compose_file" down --volumes --remove-orphans
    else
        docker compose -f "$compose_file" down --volumes --remove-orphans
    fi
    
    if [ $? -eq 0 ]; then
        log_test_pass "Services stopped successfully"
        return 0
    else
        log_test_fail "Failed to stop services"
        return 1
    fi
}

# Note: Functions are available after sourcing this file
# No explicit exports needed when sourcing
