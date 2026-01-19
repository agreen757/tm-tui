# Windows Installation Testing

This directory contains Docker-based testing infrastructure for validating Task Master TUI installation on Windows/Git Bash environments.

## Overview

Since Windows containers are not available on macOS Docker, we simulate a Git Bash environment using Ubuntu with appropriate path handling and npm configuration.

## Files

- `Dockerfile.windows-test` - Docker image that simulates Git Bash/Windows environment
- `test-installation.sh` - Comprehensive test script for installation workflow
- `README.md` - This file

## Prerequisites

- Docker installed and running
- Task Master TUI source code (parent directory)

## Running Tests

### Option 1: Automated Test Run

```bash
# From the project root
./tests/windows/run-tests.sh
```

### Option 2: Manual Testing

```bash
# Build the Docker image
docker build -f tests/windows/Dockerfile.windows-test -t tm-tui-windows-test .

# Run the container with project mounted
docker run -it --rm \
  -v "$(pwd)":/home/testuser/workspace \
  tm-tui-windows-test \
  /home/testuser/workspace/tests/windows/test-installation.sh
```

### Option 3: Interactive Testing

```bash
# Start an interactive shell in the container
docker run -it --rm \
  -v "$(pwd)":/home/testuser/workspace \
  tm-tui-windows-test \
  /bin/bash

# Inside the container, run tests manually
cd /home/testuser/workspace
./tests/windows/test-installation.sh
```

## Test Coverage

The test suite validates:

1. **Environment Setup**
   - Node.js and npm installation
   - Make availability
   - npm global directory configuration

2. **PATH Configuration**
   - npm bin directory in PATH
   - Windows-style path handling simulation

3. **Installation Workflow**
   - `make install-task-master` execution
   - task-master binary accessibility
   - Version verification

4. **Idempotency**
   - Repeated installation succeeds
   - No conflicts or errors on re-install

5. **Binary Location**
   - task-master in expected npm bin directory
   - Correct permissions and executability

## Expected Results

All tests should pass with `✅ PASS` indicators. Any `❌ FAIL` indicates an issue with the installation workflow that needs to be addressed.

## Windows-Specific Considerations

### Git Bash PATH Handling

Git Bash on Windows uses Unix-style paths internally (`/c/Users/...`) but interacts with Windows paths (`C:\Users\...`). The test environment simulates this behavior using:

- npm global directory at `~/.npm-global`
- PATH includes `~/.npm-global/bin`
- Unix-style path resolution

### PowerShell vs Git Bash

This test suite focuses on Git Bash behavior. For PowerShell-specific testing:

- PATH format: `C:\Users\<user>\.npm-global\bin`
- PATH separator: `;` instead of `:`
- Different environment variable syntax

### Known Issues

Document any Windows-specific issues discovered during testing:

- [ ] Issue 1: Description
- [ ] Issue 2: Description

## Troubleshooting

### Docker Build Fails

```bash
# Clean Docker cache and rebuild
docker system prune -f
docker build --no-cache -f tests/windows/Dockerfile.windows-test -t tm-tui-windows-test .
```

### npm Install Fails

```bash
# Inside container, check npm configuration
npm config list
npm config get prefix

# Verify permissions
ls -la ~/.npm-global
```

### task-master Not Found

```bash
# Check PATH
echo $PATH

# Verify installation location
npm list -g --depth=0
which task-master
```

## CI/CD Integration

To integrate these tests into CI/CD:

```yaml
# Example GitHub Actions workflow
- name: Test Windows Installation
  run: |
    docker build -f tests/windows/Dockerfile.windows-test -t tm-tui-windows-test .
    docker run --rm -v "$(pwd)":/home/testuser/workspace tm-tui-windows-test \
      /home/testuser/workspace/tests/windows/test-installation.sh
```

## Future Enhancements

- [ ] Add PowerShell-specific testing
- [ ] Test with actual Windows containers (when available)
- [ ] Add tests for spaces in paths
- [ ] Add tests for special characters in paths
- [ ] Test different Node.js versions
