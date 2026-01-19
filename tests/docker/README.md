# Docker Test Environment Setup

## Overview

This directory contains comprehensive Docker-based test environment setup scripts for Task Master TUI. These scripts automate the creation and configuration of multi-platform test containers, enabling consistent and reproducible testing across different operating systems and Node.js installation methods.

## Features

- **Multi-platform support**: Ubuntu, Alpine Linux, macOS (via Docker Desktop), Windows Server/Git Bash
- **Multiple Node.js installation methods**: apt, apk, nvm, Homebrew
- **Automated orchestration**: docker-compose configuration for running all tests
- **Idempotent setup**: Scripts can be run multiple times safely
- **Common helper functions**: Reusable utilities for Docker operations
- **Comprehensive logging**: Detailed output for debugging and CI integration

## Quick Start

### Prerequisites

- Docker Desktop or Docker Engine installed and running
- Sufficient disk space (~3-4 GB for all images)
- Network access for downloading base images and packages

### Setup All Environments

```bash
cd tests/docker
./setup-all.sh
```

This will:
1. Check Docker availability
2. Build all test images
3. Configure docker-compose
4. Provide usage instructions

### Run All Tests

```bash
cd tests/docker
docker-compose up
```

View individual test results as they run in real-time.

## Platform-Specific Setup

### Ubuntu

```bash
./setup-ubuntu.sh
```

Creates two Ubuntu 22.04 test environments:
- **ubuntu-apt**: Node.js installed via `apt install nodejs npm`
- **ubuntu-nvm**: Node.js installed via nvm

**What it tests**:
- System-wide npm installation
- User-level npm installation
- PATH configuration for both methods
- Permission handling

### Alpine Linux

```bash
./setup-alpine.sh
```

Creates two Alpine Linux test environments:
- **alpine-apk**: Node.js installed via `apk add nodejs npm`
- **alpine-nvm**: Node.js installed via nvm on Alpine

**What it tests**:
- Lightweight musl-based environment
- BusyBox utilities compatibility
- Alpine-specific PATH handling
- nvm with bash on Alpine

### macOS

```bash
./setup-macos.sh
```

Configures macOS testing options:
- Direct testing on macOS host
- Linux containers via Docker Desktop
- Docker Desktop configuration verification

**Note**: Docker Desktop on macOS runs Linux containers. For native macOS testing, use `tests/installation/test-local-macos.sh`.

### Windows / Git Bash

```bash
./setup-windows.sh
```

Creates Windows test environment:
- **Windows containers** (if available): Windows Server with Git Bash
- **Git Bash simulation** (fallback): Alpine-based Git Bash simulation

**What it tests**:
- Git Bash environment on Windows
- Node.js installation on Windows
- make command availability
- Path handling in Windows/Git Bash

## File Structure

```
tests/docker/
├── docker-compose.yml           # Orchestration configuration
├── setup-all.sh                 # Master setup script
├── setup-ubuntu.sh             # Ubuntu environment setup
├── setup-alpine.sh             # Alpine environment setup
├── setup-macos.sh              # macOS environment setup
├── setup-windows.sh            # Windows environment setup
├── common-helpers.sh           # Docker helper functions
├── Dockerfile.windows          # Windows Server Dockerfile
├── test-windows.bat            # Windows test script
└── README.md                   # This file
```

## Docker Compose Services

The `docker-compose.yml` defines the following services:

| Service | Platform | Node.js Method | Test Focus |
|---------|----------|----------------|-----------|
| `ubuntu-apt` | Ubuntu 22.04 | apt | System-wide installation |
| `ubuntu-nvm` | Ubuntu 22.04 | nvm | User-level installation |
| `alpine-apk` | Alpine 3.18 | apk | Lightweight environment |
| `alpine-nvm` | Alpine 3.18 | nvm | nvm on Alpine |
| `windows-gitbash` | Windows Server | Chocolatey | Windows containers (optional) |

## Running Tests

### All Tests at Once

```bash
docker-compose up
```

### Specific Platform Tests

```bash
# Ubuntu tests only
docker-compose up ubuntu-apt ubuntu-nvm

# Alpine tests only
docker-compose up alpine-apk alpine-nvm

# Windows tests (if Windows containers available)
docker-compose --profile windows up windows-gitbash
```

### Individual Test Containers

```bash
# Run Ubuntu apt test
docker run --rm -v "$(pwd)/../..:/workspace:ro" tm-test-ubuntu-apt:latest

# Run Alpine apk test
docker run --rm -v "$(pwd)/../..:/workspace:ro" tm-test-alpine-apk:latest

# Run Ubuntu nvm test
docker run --rm -v "$(pwd)/../..:/workspace:ro" tm-test-ubuntu-nvm:latest
```

### Interactive Container Shell

```bash
# Start an interactive bash shell in Ubuntu container
docker run --rm -it -v "$(pwd)/../..:/workspace" tm-test-ubuntu-apt:latest bash

# Explore Alpine container
docker run --rm -it -v "$(pwd)/../..:/workspace" tm-test-alpine-apk:latest sh
```

## Common Helper Functions

The `common-helpers.sh` script provides reusable functions:

### Docker Availability Checks

```bash
check_docker              # Verify Docker is installed and running
check_docker_compose      # Verify docker-compose availability
```

### Image Management

```bash
docker_build_image <name> <dockerfile> <context>
docker_run_test <image> <test_name> [extra_args]
```

### Container Operations

```bash
check_container_running <name>
check_container_health <name>
docker_exec_test <container> <description> <command>
wait_for_container <container> [timeout]
get_container_logs <container> [lines]
```

### Cleanup

```bash
docker_cleanup [pattern]          # Remove containers and images matching pattern
compose_down [compose_file]       # Stop and clean up compose services
```

## Test Verification

Each test environment verifies:

1. **Installation Success**
   - `make install-task-master` completes without errors
   - task-master binary is correctly installed
   - npm global packages are accessible

2. **PATH Configuration**
   - task-master is in PATH after installation
   - npm bin directory is properly configured
   - Binary location matches expected patterns

3. **Permission Handling**
   - No permission errors during installation
   - Binary has correct execute permissions
   - Works in non-root user context (where applicable)

4. **Idempotency**
   - Running installation twice succeeds
   - Second installation doesn't break first
   - No conflicts or permission errors

5. **Functionality**
   - `task-master --version` works
   - `make check-task-master` passes
   - Binary can be executed successfully

## Troubleshooting

### Docker Not Running

**Error**: `Cannot connect to the Docker daemon`

**Solution**:
```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker
```

### Permission Denied

**Error**: `permission denied while trying to connect to the Docker daemon socket`

**Solution**:
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

### Build Failures

**Error**: Network timeouts or package download failures

**Solution**:
- Check internet connection
- Retry the build
- Check Docker Hub status
- Use a different base image mirror if needed

### Out of Disk Space

**Error**: `no space left on device`

**Solution**:
```bash
# Clean up unused Docker resources
docker system prune -a

# Remove old test images
docker rmi $(docker images -q 'tm-test-*')
```

### Windows Container Issues

**Error**: `image operating system "windows" cannot be used`

**Solution**:
- Windows containers require Windows host
- Use Docker Desktop on Windows
- Switch to Windows containers in Docker Desktop settings
- Alternatively, use Git Bash simulation (Linux-based)

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Docker Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker
      uses: docker/setup-buildx-action@v2
    
    - name: Run Docker tests
      run: |
        cd tests/docker
        ./setup-all.sh
        docker-compose up --abort-on-container-exit
```

### GitLab CI Example

```yaml
docker-tests:
  image: docker:latest
  services:
    - docker:dind
  script:
    - cd tests/docker
    - ./setup-all.sh
    - docker-compose up --abort-on-container-exit
```

## Advanced Usage

### Custom Test Scenarios

Create custom test scripts by extending the base images:

```dockerfile
FROM tm-test-ubuntu-apt:latest

# Add custom test logic
COPY my-custom-test.sh /my-custom-test.sh
RUN chmod +x /my-custom-test.sh

CMD ["/my-custom-test.sh"]
```

### Debugging Failed Tests

```bash
# View logs from a failed container
docker logs <container_id>

# Run container interactively
docker run --rm -it -v "$(pwd)/../..:/workspace" tm-test-ubuntu-apt:latest bash

# Inspect container without auto-removal
docker-compose up ubuntu-apt
docker exec -it tm-test-ubuntu-apt bash
```

### Performance Testing

```bash
# Time test execution
time docker run --rm -v "$(pwd)/../..:/workspace:ro" tm-test-ubuntu-apt:latest

# Monitor resource usage
docker stats
```

## Maintenance

### Updating Base Images

```bash
# Pull latest base images
docker pull ubuntu:22.04
docker pull alpine:3.18

# Rebuild with --no-cache
docker-compose build --no-cache
```

### Cleaning Up

```bash
# Remove all test containers and images
cd tests/docker
docker-compose down
docker rmi $(docker images -q 'tm-test-*')

# Full Docker cleanup (careful!)
docker system prune -a --volumes
```

## Platform-Specific Notes

### Ubuntu
- Uses glibc standard library
- Standard GNU utilities
- apt package manager
- Larger image size (~100MB)

### Alpine Linux
- Uses musl libc (lightweight)
- BusyBox utilities (limited compared to GNU)
- apk package manager
- Smaller image size (~40MB)
- May have compatibility issues with some Node.js packages

### macOS
- Docker Desktop runs Linux VMs
- Native macOS testing requires running on host
- M1/M2 Macs use ARM architecture (may affect image builds)

### Windows
- Requires Windows host for true Windows containers
- Git Bash simulation provides Linux-based alternative
- Path handling differs significantly from Linux/macOS

## Contributing

When adding new test scenarios:

1. Create appropriate Dockerfile in this directory
2. Add setup script following existing naming convention
3. Update docker-compose.yml with new service
4. Update this README with new platform details
5. Test locally before committing
6. Document any platform-specific quirks

## Related Documentation

- [Installation Testing Guide](../installation/README.md)
- [Linux Installation Tests](../installation/LINUX_TEST_RESULTS.md)
- [Windows Testing](../windows/README.md)
- [Main README](../../README.md)

## License

MIT
