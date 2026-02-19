# CI/CD Integration Contracts

**Feature**: 006-phase-1-completion  
**Date**: 2026-02-13

---

## Overview

This document defines test contracts for CI/CD integration examples. Each example must be tested in real CI pipelines before release.

---

## Contract 1: GitHub Actions Workflow

### File: `examples/ci/github-actions.yml`

**Purpose**: Validate architecture in GitHub Actions on pull requests

**Workflow Definition**:
```yaml
name: Validate Architecture
on: [pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install loko
        run: |
          curl -L https://github.com/madstone-tech/loko/releases/latest/download/loko-linux-amd64 -o loko
          chmod +x loko
      
      - name: Validate Architecture
        run: ./loko validate --strict --exit-code
      
      - name: Build Documentation
        run: ./loko build --format html
      
      - name: Upload Artifacts
        uses: actions/upload-artifact@v4
        with:
          name: architecture-docs
          path: dist/
```

### Test Scenario 1: Valid Architecture

**Given**: PR with valid architecture (no errors, no warnings)

**When**: Workflow executes

**Then**:
- All steps complete successfully
- Workflow status: ✅ Success
- Artifacts uploaded with HTML documentation
- Total execution time: < 2 minutes

---

### Test Scenario 2: Architecture with Errors

**Given**: PR with orphaned references (validation error)

**When**: Workflow executes

**Then**:
- Step "Validate Architecture" fails with exit code 1
- Workflow status: ❌ Failure
- Error message displayed in workflow log
- Subsequent steps (Build Documentation, Upload Artifacts) skipped

**Expected Error Output**:
```
❌ Validation Errors:
  - Orphaned reference: system 'payment-service' references non-existent container 'old-api'

Validation failed with 1 error
```

---

### Test Scenario 3: Architecture with Warnings (Strict Mode)

**Given**: PR with warnings (missing description, no errors)

**When**: Workflow executes with `--strict` flag

**Then**:
- Step "Validate Architecture" fails (warnings treated as errors)
- Workflow status: ❌ Failure
- Warning message displayed as error
- Clear indication that `--strict` mode is active

**Expected Output**:
```
⚠ Warnings (treated as errors in strict mode):
  - Container 'api-gateway' has no description

Validation failed with 1 warning (strict mode enabled)
```

---

### Test Scenario 4: loko Installation Failure

**Given**: loko binary download URL returns 404

**When**: Workflow executes

**Then**:
- Step "Install loko" fails
- Workflow status: ❌ Failure
- Clear error message about download failure
- Troubleshooting guidance in logs

---

## Contract 2: GitLab CI Pipeline

### File: `examples/ci/gitlab-ci.yml`

**Purpose**: Validate architecture in GitLab CI with Docker

**Pipeline Definition**:
```yaml
validate-architecture:
  stage: test
  image: ghcr.io/madstone-tech/loko:latest
  script:
    - loko validate --strict --exit-code
    - loko build --format html
  artifacts:
    paths:
      - dist/
    expire_in: 30 days
```

### Test Scenario 1: Valid Architecture

**Given**: Commit with valid architecture

**When**: Pipeline executes

**Then**:
- Job "validate-architecture" succeeds
- Pipeline status: ✅ Passed
- Artifacts (dist/) available for 30 days
- Total execution time: < 1 minute (Docker image cached)

---

### Test Scenario 2: Architecture with Errors

**Given**: Commit with C4 hierarchy violation

**When**: Pipeline executes

**Then**:
- Job "validate-architecture" fails
- Pipeline status: ❌ Failed
- Error message in job log
- Artifacts NOT created (build step not reached)

---

### Test Scenario 3: Docker Image Not Found

**Given**: GitLab CI attempts to pull `ghcr.io/madstone-tech/loko:latest` but image doesn't exist

**When**: Pipeline executes

**Then**:
- Job fails at image pull stage
- Pipeline status: ❌ Failed
- Error: `Error response from daemon: manifest for ghcr.io/madstone-tech/loko:latest not found`

---

## Contract 3: Docker Compose Dev Environment

### File: `examples/docker-compose.dev.yml`

**Purpose**: Local development with watch mode and hot reload

**Compose Definition**:
```yaml
version: '3.8'
services:
  loko:
    image: ghcr.io/madstone-tech/loko:latest
    volumes:
      - .:/workspace
    working_dir: /workspace
    command: watch
    ports:
      - "8080:8080"
```

### Test Scenario 1: Initial Start

**Given**: Project directory with valid architecture

**When**: `docker-compose -f examples/docker-compose.dev.yml up`

**Then**:
- Container starts successfully
- loko watch mode activates
- Documentation server available at http://localhost:8080
- Logs show: `✓ Serving at http://localhost:8080`

---

### Test Scenario 2: File Change Detection

**Given**: loko container running in watch mode

**When**: User edits `src/System/system.md`

**Then**:
- File change detected within 100ms
- Rebuild triggered automatically
- Browser auto-refreshes (hot reload)
- Rebuild completes within 500ms
- Logs show: `✓ Rebuilt in 342ms`

---

### Test Scenario 3: Volume Mount

**Given**: loko container running

**When**: User creates new file `src/NewSystem/system.md` on host

**Then**:
- File appears in container at `/workspace/src/NewSystem/system.md`
- Watch mode detects new file
- Rebuild includes new system
- Documentation updated with new system

---

## Integration Test: End-to-End CI Workflow

**Goal**: Test full PR workflow with architecture changes

**Steps**:
1. Create feature branch with architecture changes
2. Push to trigger CI pipeline
3. Verify validation runs and reports errors
4. Fix errors, push again
5. Verify validation passes, artifacts created
6. Merge PR

**Test Implementation**:
```bash
# Create test branch
git checkout -b test-ci-workflow

# Add intentional error (orphaned reference)
echo "depends_on: nonexistent-container" >> src/System/container.md

# Push to trigger CI
git add . && git commit -m "test: CI validation" && git push origin test-ci-workflow

# Wait for CI (automated check)
gh pr checks --watch

# Expected: CI fails with clear error message

# Fix error
git revert HEAD
git push origin test-ci-workflow

# Expected: CI passes, artifacts uploaded
```

---

## Performance Contracts

### GitHub Actions

**Target**: < 2 minutes total execution time

**Breakdown**:
- Checkout code: < 10s
- Install loko: < 20s
- Validate: < 30s
- Build: < 60s
- Upload artifacts: < 10s

**Total**: ~130s < 2 minutes ✅

---

### GitLab CI

**Target**: < 1 minute total execution time (with cached Docker image)

**Breakdown**:
- Pull Docker image: < 5s (cached)
- Validate: < 20s
- Build: < 30s
- Upload artifacts: < 5s

**Total**: ~60s = 1 minute ✅

---

### Docker Compose

**Target**: < 5 seconds startup, < 500ms rebuild

**Metrics**:
- Container start: < 2s
- Initial build: < 3s
- File change → rebuild: < 500ms
- Browser refresh latency: < 100ms

---

## Error Message Contracts

### Validation Failure (Exit Code 1)

**Format**:
```
❌ Validation Errors:
  - [Error 1 description with file path]
  - [Error 2 description with file path]

Validation failed with N error(s)
Exit code: 1
```

**Requirements**:
- Clear ❌ indicator
- Each error on separate line with file path
- Summary line with error count
- Exit code explicitly shown (for CI debugging)

---

### Validation Warning (Strict Mode)

**Format**:
```
⚠ Warnings (treated as errors in strict mode):
  - [Warning 1 description with file path]
  - [Warning 2 description with file path]

Validation failed with N warning(s) (strict mode enabled)
Exit code: 1
```

**Requirements**:
- Clear ⚠ indicator
- Explicit mention of strict mode
- Exit code 1 (same as errors)

---

### Build Success with PDF Skip

**Format**:
```
✓ Built HTML documentation (23 files)
✓ Built Markdown documentation (README.md)
⚠ Skipped PDF (veve-cli not found)
  Install: https://github.com/madstone-tech/veve-cli

Build completed with warnings
Exit code: 0
```

**Requirements**:
- ✓ for successful formats
- ⚠ for skipped PDF
- Installation link provided
- Exit code 0 (warnings don't fail build)

---

## Validation Checklist

Before release, verify:

- [ ] GitHub Actions workflow tested in real repository
- [ ] GitLab CI pipeline tested in real project
- [ ] Docker Compose tested on Linux and macOS
- [ ] All error scenarios trigger expected failures
- [ ] All success scenarios complete within time limits
- [ ] Artifacts uploaded correctly
- [ ] Error messages are clear and actionable
- [ ] Documentation includes troubleshooting guide

---

## CI Configuration Files

### Location in Repository

```
examples/
└── ci/
    ├── github-actions.yml      # GitHub Actions workflow
    ├── gitlab-ci.yml           # GitLab CI pipeline
    ├── docker-compose.dev.yml  # Docker Compose for local dev
    └── README.md               # Usage instructions
```

### README Contents

Must include:
1. **GitHub Actions Setup**: Copy workflow to `.github/workflows/`
2. **GitLab CI Setup**: Copy pipeline to `.gitlab-ci.yml`
3. **Docker Compose Setup**: Run `docker-compose -f examples/ci/docker-compose.dev.yml up`
4. **Troubleshooting**: Common issues and solutions
5. **Customization**: How to modify for specific needs

---

## Summary

**Total Test Contracts**: 15
- 4 scenarios for GitHub Actions
- 3 scenarios for GitLab CI
- 3 scenarios for Docker Compose
- 1 integration test
- 3 performance contracts
- 1 validation checklist

**Critical Success Factors**:
- ✅ Examples work first-try on free tiers
- ✅ Error messages are clear and actionable
- ✅ Performance targets met in real CI
- ✅ Troubleshooting guide comprehensive

**Next**: Test all examples in real CI/CD environments before release
