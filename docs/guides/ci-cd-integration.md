# CI/CD Integration Guide

This guide shows you how to integrate loko architecture validation into your CI/CD pipelines. By validating architecture in CI, you ensure that documentation stays consistent and errors are caught before they reach production.

## Table of Contents

- [Overview](#overview)
- [Validation Flags](#validation-flags)
- [GitHub Actions](#github-actions)
- [GitLab CI](#gitlab-ci)
- [Docker Compose (Local Development)](#docker-compose-local-development)
- [Generic Docker Usage](#generic-docker-usage)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

## Overview

loko provides two key flags for CI/CD integration:

- `--strict`: Treats warnings as errors (fails validation if warnings exist)
- `--exit-code`: Returns non-zero exit code on validation failures (required for CI)

These flags ensure that your CI pipeline fails when architecture validation issues are detected, preventing broken documentation from being merged.

## Validation Flags

### `--strict`

Treats all warnings as errors. This is recommended for CI environments to maintain high documentation quality.

```bash
loko validate --strict
```

**Without `--strict`:**
- Errors cause validation failure ❌
- Warnings are logged but don't fail validation ⚠️

**With `--strict`:**
- Errors cause validation failure ❌
- Warnings are treated as errors and cause validation failure ❌

### `--exit-code`

Returns exit code 1 when validation fails (errors or strict-mode warnings). Required for CI pipelines to detect failures.

```bash
loko validate --exit-code
```

**Without `--exit-code`:**
- Validation results printed
- Always exits with code 0 (success)

**With `--exit-code`:**
- Validation results printed
- Exits with code 1 on failure
- Exits with code 0 on success

### Combined Usage (Recommended for CI)

```bash
loko validate --strict --exit-code
```

This combination ensures:
1. Warnings are treated as errors (`--strict`)
2. Pipeline fails on any issues (`--exit-code`)

## GitHub Actions

### Quick Setup

1. Copy the example workflow:

```bash
mkdir -p .github/workflows
cp examples/ci/github-actions.yml .github/workflows/loko-validate.yml
```

2. Commit and push:

```bash
git add .github/workflows/loko-validate.yml
git commit -m "Add loko architecture validation"
git push
```

3. Create a pull request to test the workflow.

### Example Workflow

See [`examples/ci/github-actions.yml`](../../examples/ci/github-actions.yml) for the complete workflow.

**Key features:**
- Runs on pull requests to `main`/`master`
- Only triggers on changes to `src/`, `loko.toml`
- Installs loko and D2 automatically
- Validates with `--strict --exit-code`
- Uploads HTML documentation as artifacts (30-day retention)
- Comments on PR when validation fails

**Expected execution time:** 1-2 minutes

### Workflow Behavior

| Scenario | Result |
|----------|--------|
| ✅ Valid architecture | All steps pass, artifacts uploaded |
| ❌ Validation errors | Validate step fails, subsequent steps skipped |
| ⚠️ Warnings (strict mode) | Validate step fails (warnings = errors) |

### Customization

**Change trigger paths:**
```yaml
paths:
  - 'architecture/**/*.md'  # Custom architecture directory
  - 'docs/**/*.d2'          # Custom diagram directory
  - 'loko.toml'
```

**Add PDF generation:**
```yaml
- name: Install veve-cli
  run: |
    curl -L https://github.com/terrastruct/veve/releases/latest/download/veve-linux-amd64 -o /usr/local/bin/veve-cli
    chmod +x /usr/local/bin/veve-cli

- name: Build PDF Documentation
  run: |
    loko build --format pdf
```

## GitLab CI

### Quick Setup

1. Copy the example pipeline:

```bash
cp examples/ci/.gitlab-ci.yml .gitlab-ci.yml
```

2. Commit and push:

```bash
git add .gitlab-ci.yml
git commit -m "Add loko architecture validation"
git push
```

3. Create a merge request to test the pipeline.

### Example Pipeline

See [`examples/ci/.gitlab-ci.yml`](../../examples/ci/.gitlab-ci.yml) for the complete pipeline.

**Key features:**
- Three stages: validate → build → deploy
- Caches loko binary for faster subsequent runs
- Validates with `--strict --exit-code`
- Builds HTML and Markdown documentation
- Uploads artifacts (30-day expiration)
- Optional GitLab Pages deployment

**Expected execution time:** 2-3 minutes (first run), 1-2 minutes (cached)

### Pipeline Behavior

| Scenario | Result |
|----------|--------|
| ✅ Valid architecture | All stages pass, artifacts uploaded |
| ❌ Validation errors | Validate stage fails, subsequent stages skipped |
| ⚠️ Warnings (strict mode) | Validate stage fails (warnings = errors) |

### Customization

**Enable GitLab Pages:**

The example includes an optional `pages` job that deploys documentation to GitLab Pages. It's already configured - just ensure your project has Pages enabled.

**Change trigger rules:**
```yaml
rules:
  - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    changes:
      - architecture/**/*.md  # Custom path
      - docs/**/*.d2          # Custom path
      - loko.toml
```

**Adjust cache duration:**
```yaml
cache:
  key: loko-$LOKO_VERSION-$CI_COMMIT_REF_SLUG
  paths:
    - .loko-cache/
  policy: pull-push
```

## Docker Compose (Local Development)

Docker Compose provides a local development environment with watch mode - documentation rebuilds automatically when you edit files.

### Quick Setup

1. Navigate to examples:

```bash
cd examples/ci
```

2. Start the environment:

```bash
docker-compose up
```

3. Open http://localhost:8080 in your browser to view documentation.

4. Edit your architecture files in `src/` - documentation rebuilds automatically (< 500ms).

### What's Included

See [`examples/ci/docker-compose.yml`](../../examples/ci/docker-compose.yml) for the complete configuration.

**Services:**
- `loko-dev`: Watch mode service that rebuilds on file changes
- `docs-server`: Nginx server at http://localhost:8080

**Features:**
- Auto-rebuild on file changes (< 500ms)
- Volume mounts for live editing (no container rebuild needed)
- Pre-installed loko, D2, and veve-cli
- Resource limits (configurable)

### Expected Behavior

| Event | Result |
|-------|--------|
| File change detected | Rebuild triggered within 2 seconds |
| Valid change | Build completes in < 500ms, docs updated |
| Invalid change | Error logged, no dist/ update |

### Viewing Logs

```bash
# Watch build logs
docker-compose logs -f loko-dev

# Watch HTTP server logs
docker-compose logs -f docs-server
```

### Stopping

```bash
docker-compose down
```

### Clean Rebuild

```bash
docker-compose down -v
docker-compose build --no-cache
docker-compose up
```

## Generic Docker Usage

Use the loko Docker image in any CI/CD system or local environment.

### Building the Image

```bash
docker build -t loko:latest -f examples/ci/Dockerfile .
```

### Usage Examples

**Validate architecture:**
```bash
docker run --rm -v $(pwd):/workspace loko:latest validate --strict --exit-code
```

**Build HTML documentation:**
```bash
docker run --rm -v $(pwd):/workspace loko:latest build --format html
```

**Build PDF documentation:**
```bash
docker run --rm -v $(pwd):/workspace loko:latest build --format pdf
```

**Interactive shell:**
```bash
docker run --rm -it -v $(pwd):/workspace loko:latest /bin/bash
```

### What's Included

The Docker image includes:
- loko binary
- D2 diagram renderer
- veve-cli for PDF generation (optional dependency)
- All dependencies for full builds

**Expected image size:** ~250MB

## Troubleshooting

### Common Issues

#### 1. Validation Fails with "File Not Found"

**Problem:** loko can't find `loko.toml` or `src/` directory.

**Solution:** Ensure you're running loko from your project root:

```bash
# Check current directory
ls -la loko.toml src/

# If files missing, navigate to project root
cd /path/to/your/project
loko validate --strict --exit-code
```

#### 2. CI Workflow Doesn't Trigger

**Problem:** GitHub Actions or GitLab CI doesn't run on push/PR.

**Solution (GitHub Actions):** Check trigger paths match your project structure:

```yaml
paths:
  - 'src/**/*.md'      # Does your project use src/?
  - 'architecture/**/*.md'  # Or a different directory?
  - 'loko.toml'
```

**Solution (GitLab CI):** Verify rules configuration:

```yaml
rules:
  - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    changes:
      - src/**/*.md  # Match your directory structure
```

#### 3. Docker Compose "Permission Denied"

**Problem:** Docker can't access mounted volumes.

**Solution:** Check file permissions:

```bash
# Make directories readable
chmod -R 755 src/ loko.toml

# On Linux, may need to match container UID/GID
sudo chown -R $(id -u):$(id -g) src/ dist/
```

#### 4. Build Succeeds but Artifacts Missing

**Problem:** CI uploads empty `dist/` directory.

**Solution:** Ensure build runs before artifact upload:

```yaml
# GitHub Actions
- name: Build Documentation
  if: success()  # Only if validation passed
  run: loko build --format html

- name: Upload Artifacts
  if: success()  # Only if build passed
  uses: actions/upload-artifact@v4
  with:
    path: dist/  # Must exist after build
```

#### 5. Warnings Not Treated as Errors

**Problem:** CI passes despite warnings.

**Solution:** Ensure `--strict` flag is present:

```bash
# ❌ Wrong - warnings don't fail
loko validate --exit-code

# ✅ Correct - warnings treated as errors
loko validate --strict --exit-code
```

#### 6. Exit Code Always 0 (Success)

**Problem:** CI never fails even with validation errors.

**Solution:** Ensure `--exit-code` flag is present:

```bash
# ❌ Wrong - always exits 0
loko validate --strict

# ✅ Correct - exits 1 on failure
loko validate --strict --exit-code
```

### Getting Help

If you encounter issues not covered here:

1. Check the [loko GitHub Issues](https://github.com/madstone-tech/loko/issues)
2. Review CI workflow logs for detailed error messages
3. Run locally with `--verbose` flag for debugging:
   ```bash
   loko validate --strict --exit-code --verbose
   ```

## Best Practices

### 1. Always Use Both Flags in CI

```bash
loko validate --strict --exit-code
```

This ensures maximum quality and proper failure detection.

### 2. Validate Before Building

```yaml
# GitHub Actions example
- name: Validate Architecture
  run: loko validate --strict --exit-code

- name: Build Documentation
  if: success()  # Only if validation passed
  run: loko build --format html
```

Saves CI time by failing fast on validation errors.

### 3. Cache Dependencies

**GitHub Actions:**
```yaml
- name: Cache loko
  uses: actions/cache@v4
  with:
    path: ~/.loko-cache
    key: loko-${{ runner.os }}-latest
```

**GitLab CI:**
```yaml
cache:
  key: loko-$LOKO_VERSION
  paths:
    - .loko-cache/
```

Speeds up subsequent pipeline runs.

### 4. Limit Triggers to Architecture Changes

```yaml
# GitHub Actions
paths:
  - 'src/**/*.md'
  - 'src/**/*.d2'
  - 'loko.toml'

# GitLab CI
rules:
  - changes:
      - src/**/*.md
      - src/**/*.d2
      - loko.toml
```

Prevents unnecessary pipeline runs on unrelated changes.

### 5. Upload Artifacts for Review

```yaml
# GitHub Actions
- name: Upload Documentation
  uses: actions/upload-artifact@v4
  with:
    name: architecture-docs
    path: dist/
    retention-days: 30
```

Allows reviewers to preview documentation changes before merge.

### 6. Set Resource Limits (Docker Compose)

```yaml
deploy:
  resources:
    limits:
      cpus: '1.0'
      memory: 512M
```

Prevents runaway processes in local development.

### 7. Use Docker for Consistency

```bash
# Same environment everywhere
docker run --rm -v $(pwd):/workspace loko:latest validate --strict --exit-code
```

Eliminates "works on my machine" issues.

### 8. Document Custom Workflows

If you customize the examples, document changes in your project's README:

```markdown
## Architecture Validation

We use loko to validate architecture documentation in CI.

- **GitHub Actions**: See `.github/workflows/loko-validate.yml`
- **Custom paths**: Architecture lives in `docs/architecture/`
- **Custom flags**: We use `--strict --exit-code --format toon`
```

## Next Steps

- **MCP Integration**: See [MCP Integration Guide](./mcp-integration-guide.md) for AI-assisted architecture workflows
- **TOON Format**: See [TOON Format Guide](./toon-format-guide.md) for token-efficient exports
- **Examples**: Explore `examples/` directory for complete project examples

---

**Last updated:** 2025-02-13  
**loko version:** v0.2.0  
**Tested on:** GitHub Actions, GitLab CI, Docker v24.0+
