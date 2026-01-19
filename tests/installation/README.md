# Linux Installation Testing Guide

## Overview

This directory contains comprehensive Docker-based tests for Task Master installation on Linux distributions. Tests verify installation workflows with different Node.js installation methods and permission configurations.

## Test Matrix

### Ubuntu 22.04 LTS Tests

1. **test-apt.sh** - Node.js via apt package manager
   - Installs Node.js using `apt install nodejs npm`
   - Tests system-wide npm global installation
   - Verifies PATH configuration for system-installed Node
   - Tests permission handling

2. **test-nvm.sh** - Node.js via nvm
   - Installs nvm and Node.js LTS
   - Tests user-level npm global installation
   - Verifies PATH configuration for nvm-installed Node
   - Tests nvm environment setup

### Alpine Linux Tests

3. **test-apk.sh** - Node.js via apk package manager
   - Installs Node.js using `apk add nodejs npm`
   - Tests Alpine-specific PATH handling
   - Verifies musl compatibility
   - Tests lightweight environment

4. **test-nvm.sh** - Node.js via nvm on Alpine
   - Installs nvm on Alpine Linux
   - Tests nvm with Alpine's minimal environment
   - Verifies bash compatibility on Alpine

## Prerequisites

- Docker Desktop or Docker Engine running
- Sufficient disk space for Docker images (~2GB)
- Network access for package downloads

## Running Tests

### Run All Tests

```bash
./run-linux-tests.sh
```

This executes all four test scenarios and generates a comprehensive report.

### Run Individual Tests

```bash
# Ubuntu with apt
docker build -f ubuntu/Dockerfile.apt -t tm-test-ubuntu-apt ubuntu/
docker run --rm -v "$PWD/../..:/workspace:ro" tm-test-ubuntu-apt

# Ubuntu with nvm
docker build -f ubuntu/Dockerfile.nvm -t tm-test-ubuntu-nvm ubuntu/
docker run --rm -v "$PWD/../..:/workspace:ro" tm-test-ubuntu-nvm

# Alpine with apk
docker build -f alpine/Dockerfile.apk -t tm-test-alpine-apk alpine/
docker run --rm -v "$PWD/../..:/workspace:ro" tm-test-alpine-apk

# Alpine with nvm
docker build -f alpine/Dockerfile.nvm -t tm-test-alpine-nvm alpine/
docker run --rm -v "$PWD/../..:/workspace:ro" tm-test-alpine-nvm
```

## Test Scenarios

Each test verifies:

1. **Installation Success**
   - `make install-task-master` completes without errors
   - task-master binary is installed correctly
   - npm global packages are accessible

2. **PATH Configuration**
   - task-master is in PATH after installation
   - npm bin directory is properly configured
   - Binary location matches expected patterns

3. **Permission Handling**
   - No permission errors during installation
   - Binary has correct execute permissions
   - Works in non-root user context

4. **Idempotency**
   - Running installation twice succeeds
   - Second installation doesn't break first
   - No conflicts or permission errors

5. **Functionality**
   - `task-master --version` works
   - `make check-task-master` passes
   - Binary can be executed

## Test Results

Results are stored in `results/` directory:
- Individual test logs: `{test-name}_{timestamp}.log`
- Aggregated run log: `test-run_{timestamp}.txt`

Each log contains:
- Package installation output
- make target execution results
- PATH verification details
- Permission checks
- Error messages (if any)

## Common Issues

### Docker Not Running

Error: `Cannot connect to Docker daemon`

**Solution**: Start Docker Desktop or Docker Engine

```bash
# macOS - start Docker Desktop
open -a Docker

# Linux - start Docker service
sudo systemctl start docker
```

### Permission Errors in Container

Error: `EACCES: permission denied`

**Solution**: This is expected and should be caught by test scripts. Check test logs for proper error handling.

### npm Install Fails

Error: `npm ERR! network`

**Solution**: Check internet connectivity. Docker containers need network access to download npm packages.

### PATH Issues

If task-master is installed but not in PATH:

**Expected Behavior**: Tests document PATH configuration requirements for each platform.

## Expected Results

### Ubuntu + apt

- task-master installed to `/usr/local/bin` or `/usr/bin`
- System-wide npm global prefix
- Works without additional PATH configuration

### Ubuntu + nvm

- task-master installed to `~/.nvm/versions/node/*/bin/`
- nvm manages PATH automatically
- Requires nvm initialization in shell profile

### Alpine + apk

- task-master installed to `/usr/bin` or `/usr/local/bin`
- Alpine's minimal environment requires bash for nvm
- Works with Alpine's BusyBox utilities

### Alpine + nvm

- task-master in nvm-managed directory
- Requires bash (not sh) for nvm
- PATH managed by nvm initialization

## Platform-Specific Notes

### Ubuntu (Debian-based)

- Uses apt package manager
- Node.js from Ubuntu repositories (may be older version)
- npm global prefix: `/usr/local` or `/usr`
- Standard GNU userland

### Alpine Linux

- Uses apk package manager
- Node.js from Alpine repositories
- Lightweight (musl libc instead of glibc)
- BusyBox utilities (limited compared to GNU)
- nvm requires bash installation

## CI/CD Integration

These tests can be integrated into CI pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Linux Installation Tests
  run: ./tests/installation/run-linux-tests.sh
```

## Troubleshooting

### Test Hangs

If a test hangs:
1. Check Docker daemon status
2. Verify network connectivity
3. Check available disk space
4. Review individual test logs

### Unexpected Failures

If tests fail unexpectedly:
1. Run test individually with verbose output
2. Check test log in `results/` directory
3. Verify Makefile targets work outside Docker
4. Test with newer/older Docker image versions

## Maintenance

When updating tests:
1. Update test scripts in `ubuntu/` and `alpine/` directories
2. Update Dockerfiles if base image versions change
3. Update this README with new test scenarios
4. Test changes locally before committing
