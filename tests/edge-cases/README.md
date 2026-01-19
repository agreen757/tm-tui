# Edge Case Testing Documentation

## Overview

This directory contains Docker-based edge case testing for Task Master TUI installation. Each test simulates a specific failure scenario to verify error handling and recovery procedures.

## Test Cases

### Test 1: Missing npm
- **Scenario**: npm not installed on system
- **Expected**: Clear error message directing user to install Node.js/npm
- **Dockerfile**: `dockerfiles/1-no-npm.Dockerfile`

### Test 2: Already Installed
- **Scenario**: task-master-ai already installed globally
- **Expected**: Installation succeeds idempotently or reports "already installed"
- **Dockerfile**: `dockerfiles/2-already-installed.Dockerfile`

### Test 3: Misconfigured PATH
- **Scenario**: npm global bin directory not in PATH
- **Expected**: Clear error with instructions to fix PATH
- **Dockerfile**: `dockerfiles/3-misconfigured-path.Dockerfile`

### Test 4: Permission Errors
- **Scenario**: User lacks permissions to install global packages
- **Expected**: Clear EACCES error with fix options (npm config prefix or nvm)
- **Dockerfile**: `dockerfiles/4-permission-errors.Dockerfile`

### Test 5: Old npm Version
- **Scenario**: Older npm version (Node 14)
- **Expected**: Installation succeeds or provides compatibility warning
- **Dockerfile**: `dockerfiles/5-old-npm.Dockerfile`

### Test 6: Limited Disk Space
- **Scenario**: Insufficient disk space for installation
- **Expected**: Clear ENOSPC error
- **Dockerfile**: `dockerfiles/6-disk-space.Dockerfile`

### Test 7: Multiple Global Packages
- **Scenario**: Many existing global npm packages
- **Expected**: Installation succeeds without conflicts
- **Dockerfile**: `dockerfiles/7-multiple-packages.Dockerfile`

## Running Tests

### Run All Tests
```bash
./scripts/run-all-tests.sh
```

This runs all edge case tests in sequence and generates a detailed report in `results/test-report-<timestamp>.md`.

### Run Single Test
```bash
./scripts/run-single-test.sh <test_number>
```

Example:
```bash
./scripts/run-single-test.sh 1  # Test missing npm
./scripts/run-single-test.sh 4  # Test permission errors
```

This builds the Docker container and drops you into an interactive shell for manual testing.

### Manual Testing in Container

Once inside a test container:

```bash
# Check npm availability
which npm
npm --version

# Try to install task-master
make install-task-master

# Check if task-master is accessible
make check-task-master

# Try to use task-master
task-master --version
task-master --help
```

## Test Output

Test results are saved to `results/`:
- `test-report-<timestamp>.md` - Comprehensive test report
- `test-<N>-output.log` - stdout/stderr from test N
- `test-<N>-status.txt` - Pass/fail status for test N
- `test-<N>-build.log` - Docker build logs for test N

## Interpreting Results

### Successful Tests
- ✅ Clear, actionable error messages
- ✅ Error messages match README documentation
- ✅ Recovery procedures work as documented
- ✅ No unexpected failures or hangs

### Failed Tests
- ❌ Missing or unclear error messages
- ❌ Inconsistent documentation
- ❌ Recovery procedures don't work
- ❌ Unexpected behavior

## Adding New Tests

1. Create a new Dockerfile in `dockerfiles/`:
   ```dockerfile
   # Test Case N: Description
   FROM node:18-slim
   # Set up your test scenario
   ```

2. Add test case to `run-all-tests.sh`:
   ```bash
   run_test "N" "Test Name" \
       "tests/edge-cases/dockerfiles/N-test-name.Dockerfile" \
       "your test command here"
   ```

3. Add validation logic in the case statement:
   ```bash
   N)
       if grep -q "expected pattern" "$output_file"; then
           test_passed=true
       fi
       ;;
   ```

4. Document the test case in this README

## CI/CD Integration

To integrate with CI/CD:

```yaml
# .github/workflows/edge-case-tests.yml
name: Edge Case Tests

on: [push, pull_request]

jobs:
  edge-cases:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run edge case tests
        run: |
          cd tests/edge-cases
          ./scripts/run-all-tests.sh
      - name: Upload test results
        uses: actions/upload-artifact@v2
        if: always()
        with:
          name: edge-case-test-results
          path: tests/edge-cases/results/
```

## Prerequisites

- Docker installed and running
- Bash shell (macOS, Linux, WSL on Windows)
- Make utility

## Troubleshooting

### Docker Build Failures
- Ensure Docker daemon is running
- Check Docker has internet access for package downloads
- Verify sufficient disk space

### Test Hangs
- Some containers may wait for input; use Ctrl+C to exit
- Check container logs: `docker logs <container_id>`

### Permission Issues
- On Linux, you may need to run Docker commands with sudo
- Or add your user to the docker group: `sudo usermod -aG docker $USER`

## Cross-Platform Considerations

These tests use Linux-based Docker containers. For platform-specific testing:
- **macOS**: Test on macOS host directly
- **Windows**: Test in WSL2 environment or Windows containers
- **Linux**: These Docker tests should work natively

## Next Steps

1. Run all tests: `./scripts/run-all-tests.sh`
2. Review test report in `results/`
3. Document findings in project issues or documentation
4. Update README.md with improved troubleshooting steps
5. Fix any gaps in error handling or documentation
