# Edge Case Testing Quick Reference

## Quick Start

```bash
# Navigate to test directory
cd tests/edge-cases

# Validate framework setup
./scripts/validate-framework.sh

# Run all tests (generates report)
./scripts/run-all-tests.sh

# View latest report
cat results/test-report-*.md | tail -100
```

## Individual Test Commands

```bash
# Test 1: Missing npm
./scripts/run-single-test.sh 1
# Inside container: make check-task-master

# Test 2: Already installed
./scripts/run-single-test.sh 2
# Inside container: make check-task-master && task-master --version

# Test 3: Misconfigured PATH
./scripts/run-single-test.sh 3
# Inside container: export PATH=/usr/bin:/bin && make check-task-master

# Test 4: Permission errors
./scripts/run-single-test.sh 4
# Inside container: npm install -g task-master-ai

# Test 5: Old npm version
./scripts/run-single-test.sh 5
# Inside container: npm --version && make install-task-master

# Test 6: Disk space
./scripts/run-single-test.sh 6
# Inside container: df -h && make install-task-master

# Test 7: Multiple packages
./scripts/run-single-test.sh 7
# Inside container: npm list -g --depth=0 && make install-task-master
```

## Test Results Location

```
tests/edge-cases/results/
├── test-report-<timestamp>.md       # Main report
├── test-1-output.log                # Test 1 stdout/stderr
├── test-1-status.txt                # Test 1 pass/fail
├── test-1-build.log                 # Test 1 Docker build log
└── ... (similar for tests 2-7)
```

## Interpreting Results

### Pass Indicators
- ✅ Expected error messages present
- ✅ Error messages match documentation
- ✅ Recovery procedures work
- ✅ No unexpected failures

### Fail Indicators
- ❌ Missing error messages
- ❌ Unclear user guidance
- ❌ Documentation mismatch
- ❌ Unexpected behavior

## Common Manual Test Commands

Once inside a test container:

```bash
# Check npm status
which npm
npm --version
npm config get prefix

# Check task-master status
which task-master
task-master --version

# Try installation
make install-task-master

# Verify installation
make check-task-master

# Check PATH
echo $PATH
ls -la $(npm config get prefix)/bin/

# Test with different PATH
export PATH=/usr/bin:/bin
make check-task-master
```

## Adding a New Test

1. Create Dockerfile: `dockerfiles/8-new-test.Dockerfile`
2. Add to `scripts/run-all-tests.sh`:
   ```bash
   run_test "8" "New Test Name" \
       "tests/edge-cases/dockerfiles/8-new-test.Dockerfile" \
       "test command here"
   ```
3. Add validation logic in case statement
4. Document in README.md

## CI/CD Integration

```yaml
# .github/workflows/edge-cases.yml
name: Edge Case Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - run: cd tests/edge-cases && ./scripts/run-all-tests.sh
      - uses: actions/upload-artifact@v2
        with:
          name: test-results
          path: tests/edge-cases/results/
```

## Troubleshooting

### Docker not found
```bash
# Install Docker
# macOS: brew install docker
# Linux: apt-get install docker.io
```

### Permission denied (Docker)
```bash
# Linux: Add user to docker group
sudo usermod -aG docker $USER
# Then logout/login
```

### Build fails
```bash
# Check Docker daemon
docker ps

# Check internet connectivity
ping google.com

# Check disk space
df -h
```

### Test hangs
- Use Ctrl+C to stop
- Check container status: `docker ps`
- Kill container: `docker stop <container_id>`

## Best Practices

1. **Run tests before releases** to catch installation regressions
2. **Update tests** when adding new Makefile targets
3. **Document findings** in FINDINGS.md
4. **Keep Dockerfiles minimal** for faster builds
5. **Test on real systems** for platform-specific issues

## Files Overview

```
tests/edge-cases/
├── dockerfiles/              # Test container definitions
│   ├── 1-no-npm.Dockerfile
│   ├── 2-already-installed.Dockerfile
│   ├── 3-misconfigured-path.Dockerfile
│   ├── 4-permission-errors.Dockerfile
│   ├── 5-old-npm.Dockerfile
│   ├── 6-disk-space.Dockerfile
│   └── 7-multiple-packages.Dockerfile
├── scripts/                  # Test automation
│   ├── run-all-tests.sh     # Automated test suite
│   ├── run-single-test.sh   # Manual single test
│   └── validate-framework.sh # Framework validation
├── results/                  # Test output (generated)
│   └── test-report-*.md
├── README.md                 # Full documentation
├── FINDINGS.md              # Detailed findings and recommendations
├── GITHUB_ISSUE_TEMPLATE.md # Issue template
└── QUICK_REFERENCE.md       # This file
```

## Expected Test Times

- Framework validation: ~30 seconds
- Single test: ~1-2 minutes
- Full suite: ~10-15 minutes

Times vary based on Docker image caching and network speed.

## Support

For questions or issues with the test framework:
1. Check README.md for detailed documentation
2. Review FINDINGS.md for known issues and recommendations
3. Run validate-framework.sh to check setup
4. Open GitHub issue with test output logs

---

**Last Updated:** 2026-01-19
**Framework Version:** 1.0
