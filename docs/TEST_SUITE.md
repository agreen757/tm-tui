# Test Suite Documentation

## Overview

The Task Master TUI project uses a comprehensive test suite with coverage reporting and CI/CD integration. The test infrastructure is built using `github.com/stretchr/testify/suite` for structured testing with consistent setup and teardown.

## Test Suite Structure

### ExecutionWorkflowTestSuite

Located at `internal/executor/execution_workflow_suite_test.go`, this test suite provides structured testing for the executor package.

**Key Features:**
- Uses testify's suite functionality
- Consistent SetupTest/TearDownTest lifecycle hooks
- Comprehensive coverage of execution workflows
- Real-time output streaming tests
- Concurrent execution handling
- Cancellation and cleanup verification

**Example Usage:**

```go
type ExecutionWorkflowTestSuite struct {
    suite.Suite
    service   *Service
    tmpDir    string
    tmDir     string
    logPath   string
    cfg       *config.Config
}

func (s *ExecutionWorkflowTestSuite) SetupTest() {
    // Create temp directories and initialize service
    tmpDir, err := os.MkdirTemp("", "executor-suite-test-*")
    s.Require().NoError(err)
    // ... setup code
}

func (s *ExecutionWorkflowTestSuite) TearDownTest() {
    // Clean up resources after each test
    if s.service != nil {
        s.service.Close()
    }
    os.RemoveAll(s.tmpDir)
}

func TestExecutionWorkflowTestSuite(t *testing.T) {
    suite.Run(t, new(ExecutionWorkflowTestSuite))
}
```

## Coverage Reporting

### Local Development

**Run tests with coverage:**
```bash
make test-coverage
```

**Generate HTML coverage report:**
```bash
make coverage-html
open coverage.html
```

**Run full test suite with threshold verification:**
```bash
make test-suite
```

**Coverage output example:**
```
Running tests with coverage...
ok      github.com/agreen757/tm-tui/internal/executor   33.530s coverage: 85.8% of statements
ok      github.com/agreen757/tm-tui/internal/git        17.579s coverage: 90.5% of statements
ok      github.com/agreen757/tm-tui/internal/pathutil    1.864s coverage: 92.9% of statements

Coverage Summary:
total:                                                  (statements)    85.8%
```

### Coverage Threshold

The project targets **>80% code coverage** across core components.

**Verify coverage meets threshold:**
```bash
./scripts/check-coverage.sh coverage.out 80
```

**Output:**
```
Current coverage: 85.8%
Required threshold: 80%
✓ Coverage meets threshold
```

### Package-Specific Coverage

| Package                  | Coverage | Target | Status |
|-------------------------|----------|--------|--------|
| internal/executor       | 85.8%    | >80%   | ✓      |
| internal/git            | 90.5%    | >80%   | ✓      |
| internal/pathutil       | 92.9%    | >80%   | ✓      |
| internal/types          | 100.0%   | >80%   | ✓      |
| internal/config         | 72.3%    | >80%   | 🔄     |

## CI/CD Integration

### GitHub Actions Workflow

File: `.github/workflows/test.yml`

**Triggers:**
- Push to `main`, `develop`, `concurrent-task-execution` branches
- Pull requests to `main`, `develop`

**Matrix Testing:**
- **Operating Systems:** Ubuntu Latest, macOS Latest
- **Go Versions:** 1.23.x, 1.24.x

**Pipeline Steps:**

1. **Checkout Code**
   ```yaml
   - uses: actions/checkout@v4
   ```

2. **Setup Go**
   ```yaml
   - uses: actions/setup-go@v5
     with:
       go-version: ${{ matrix.go-version }}
   ```

3. **Cache Go Modules**
   ```yaml
   - uses: actions/cache@v4
     with:
       path: |
         ~/.cache/go-build
         ~/go/pkg/mod
   ```

4. **Run Tests with Coverage**
   ```yaml
   - run: go test -v ./... -coverprofile=coverage.out -covermode=atomic -race
   ```

5. **Verify Coverage Threshold**
   ```yaml
   - run: |
       COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
       echo "Current coverage: ${COVERAGE}%"
       if (( $(echo "$COVERAGE >= 80" | bc -l) )); then
         echo "✓ Coverage meets 80% threshold"
       else
         echo "⚠ Coverage is ${COVERAGE}% (target: 80%)"
       fi
   ```

6. **Upload to Codecov**
   ```yaml
   - uses: codecov/codecov-action@v4
     with:
       file: ./coverage.out
   ```

7. **Save Coverage Artifact**
   ```yaml
   - uses: actions/upload-artifact@v4
     with:
       name: coverage-report
       path: coverage.out
   ```

### Lint Job

Separate lint job runs:
- `go fmt` format checking
- `go vet` static analysis
- `staticcheck` advanced linting

## Makefile Targets

### Test Execution

| Target              | Description                                      |
|--------------------|--------------------------------------------------|
| `make test`        | Run all tests                                    |
| `make test-coverage` | Run tests with coverage report                 |
| `make coverage-html` | Generate HTML coverage report                  |
| `make test-unit`   | Run unit tests only (short mode)                |
| `make test-integration` | Run integration tests only                  |
| `make test-suite`  | Run full suite with threshold verification       |
| `make test-ci`     | Run tests as CI would (with race detector)       |

### Example Commands

**Basic testing:**
```bash
make test
```

**Coverage with HTML report:**
```bash
make coverage-html
```

**Full suite with verification:**
```bash
make test-suite
```

**CI simulation:**
```bash
make test-ci
```

## Test Writing Guidelines

### Using testify/suite

1. **Create a test suite struct:**
   ```go
   type MyTestSuite struct {
       suite.Suite
       // Add fields for test fixtures
   }
   ```

2. **Implement SetupTest (optional):**
   ```go
   func (s *MyTestSuite) SetupTest() {
       // Runs before each test
   }
   ```

3. **Implement TearDownTest (optional):**
   ```go
   func (s *MyTestSuite) TearDownTest() {
       // Runs after each test
   }
   ```

4. **Write test methods:**
   ```go
   func (s *MyTestSuite) TestSomething() {
       s.Equal(expected, actual)
       s.NoError(err)
   }
   ```

5. **Run the suite:**
   ```go
   func TestMyTestSuite(t *testing.T) {
       suite.Run(t, new(MyTestSuite))
   }
   ```

### Assertions

testify provides rich assertions:

```go
s.Equal(expected, actual)
s.NotEqual(a, b)
s.Nil(value)
s.NotNil(value)
s.True(condition)
s.False(condition)
s.NoError(err)
s.Error(err)
s.Contains(haystack, needle)
s.Len(list, 3)
s.Greater(a, b)
```

### Test Coverage Best Practices

1. **Test happy paths and edge cases**
2. **Test error handling**
3. **Test concurrent operations**
4. **Test resource cleanup**
5. **Use table-driven tests for multiple scenarios**
6. **Mock external dependencies**
7. **Verify state changes**

## Coverage Tools

### View coverage by function:
```bash
go tool cover -func=coverage.out
```

### View coverage in browser:
```bash
go tool cover -html=coverage.out
```

### Filter coverage by package:
```bash
go tool cover -func=coverage.out | grep "internal/executor"
```

### Extract total coverage:
```bash
go tool cover -func=coverage.out | grep total
```

## Troubleshooting

### Tests failing locally but passing in CI
- Check Go version consistency
- Verify dependencies are up to date: `go mod download`
- Clean build cache: `go clean -cache`

### Coverage drops unexpectedly
- Run `git diff` to see which files changed
- Check for new untested code paths
- Use `go tool cover -html=coverage.out` to visualize gaps

### Race detector failures
- Review concurrent code for race conditions
- Use `go test -race` to identify specific races
- Add proper synchronization (mutexes, channels)

### CI pipeline fails
- Check GitHub Actions logs for specific errors
- Reproduce locally with `make test-ci`
- Verify all dependencies are available

## Future Improvements

- [ ] Increase coverage to 85%+ across all packages
- [ ] Add benchmark tests for performance-critical paths
- [ ] Implement property-based testing for complex logic
- [ ] Add mutation testing to verify test quality
- [ ] Create test fixtures library for common scenarios
- [ ] Add performance regression testing

## References

- [testify documentation](https://github.com/stretchr/testify)
- [Go testing package](https://pkg.go.dev/testing)
- [Go code coverage](https://go.dev/blog/cover)
- [GitHub Actions](https://docs.github.com/en/actions)
- [Codecov](https://about.codecov.io/)
