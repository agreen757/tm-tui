---
name: Installation Edge Case Improvements
about: Improvements identified from edge case testing
title: '[INSTALLATION] Improve error handling and user guidance'
labels: enhancement, installation, documentation
assignees: ''
---

## Summary

Based on comprehensive edge case testing using Docker containers, several improvements have been identified for the installation process. These changes will improve error messages, add pre-flight checks, and enhance user guidance for common installation issues.

## Background

A Docker-based edge case testing framework was created to test 7 different installation scenarios:
1. Missing npm
2. Already installed task-master
3. Misconfigured PATH
4. Permission errors
5. Old npm version
6. Limited disk space
7. Multiple global packages

Full findings: `tests/edge-cases/FINDINGS.md`

## Issues Identified

### 1. Missing npm Error Handling (High Priority)

**Current:** Generic "npm: command not found" error
**Needed:** Clear guidance on installing Node.js/npm with platform-specific instructions

### 2. PATH Misconfiguration Detection (High Priority)

**Current:** No detection of task-master installed but not in PATH
**Needed:** Detect this scenario and provide exact PATH fix commands

### 3. Permission Error Guidance (High Priority)

**Current:** Generic EACCES error from npm
**Needed:** Detect permission errors and provide fix options (npm config prefix or nvm)

### 4. Version Compatibility Checks (Medium Priority)

**Current:** No version validation
**Needed:** Check minimum Node.js/npm version and warn if outdated

## Proposed Solutions

### Makefile Enhancements

```makefile
# Add pre-flight checks
.PHONY: check-node
check-node:
	@command -v node >/dev/null 2>&1 || { \
		echo "Error: Node.js is not installed"; \
		echo "Visit: https://nodejs.org/ or install via nvm"; \
		exit 1; \
	}

# Enhanced task-master check with PATH detection
.PHONY: check-task-master
check-task-master:
	@if ! command -v task-master >/dev/null 2>&1; then \
		NPM_PREFIX=$$(npm config get prefix 2>/dev/null || echo "/usr/local"); \
		if [ -f "$$NPM_PREFIX/bin/task-master" ]; then \
			echo "task-master is installed but not in PATH."; \
			echo "Add to ~/.zshrc: export PATH=\"$$NPM_PREFIX/bin:\$$PATH\""; \
		else \
			echo "task-master not found. Run 'make install-task-master'"; \
		fi; \
		exit 1; \
	fi

# Improved installation with error detection
.PHONY: install-task-master
install-task-master: check-node
	@npm install -g task-master-ai || { \
		echo "Installation failed. See README.md#troubleshooting"; \
		exit 1; \
	}
```

### README.md Updates

1. Add Prerequisites section with version requirements
2. Enhance Troubleshooting section with:
   - PATH detection and fix
   - Permission error solutions
   - npm version compatibility
3. Add Quick Start validation steps

## Implementation Plan

1. **Week 1: Critical Fixes**
   - [ ] Enhance Makefile error detection
   - [ ] Add PATH detection to check-task-master
   - [ ] Update README troubleshooting section

2. **Week 2: Validation**
   - [ ] Add npm version checks
   - [ ] Test on all platforms (macOS, Linux, Windows/WSL)
   - [ ] Validate with edge case test suite

3. **Week 3: Automation**
   - [ ] Set up CI/CD with Docker tests
   - [ ] Create regression test suite
   - [ ] Document testing procedures

## Testing

Edge case testing framework available in `tests/edge-cases/`:

```bash
# Run all tests
./tests/edge-cases/scripts/run-all-tests.sh

# Run specific test
./tests/edge-cases/scripts/run-single-test.sh <1-7>

# Validate framework
./tests/edge-cases/scripts/validate-framework.sh
```

## Success Criteria

- [ ] All high-priority error scenarios have clear, actionable error messages
- [ ] PATH misconfiguration is automatically detected and fixed
- [ ] Permission errors provide at least 2 fix options
- [ ] README.md troubleshooting section covers all tested edge cases
- [ ] Edge case test suite passes with 100% success rate

## Related Files

- `tests/edge-cases/FINDINGS.md` - Detailed analysis and recommendations
- `tests/edge-cases/README.md` - Test framework documentation
- `Makefile` - Current implementation
- `README.md` - User-facing documentation

## Additional Context

The Docker-based testing framework can be integrated into CI/CD to prevent regressions. Each test case simulates a specific failure mode in an isolated container.

Full implementation recommendations with code examples are available in `tests/edge-cases/FINDINGS.md`.
