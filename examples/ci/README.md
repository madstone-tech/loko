# CI/CD Integration Examples

This directory contains examples for integrating loko architecture validation into CI/CD pipelines.

## Examples Included

### 1. GitHub Actions (`github-actions.yml`)
- Validates architecture on pull requests
- Uses `loko validate --strict --exit-code` to fail on warnings
- Uploads HTML documentation as artifacts
- Free tier compatible

### 2. GitLab CI (`.gitlab-ci.yml`)
- Validates architecture in merge requests
- Uploads generated documentation as pipeline artifacts
- Caches loko binary for faster builds
- Free tier compatible

### 3. Docker Compose (`docker-compose.yml`)
- Local development environment with watch mode
- Auto-rebuilds documentation on file changes (< 500ms)
- Includes loko with veve-cli pre-installed
- Volume mounts for live editing

### 4. Dockerfile (`Dockerfile`)
- Containerized loko with veve-cli for PDF generation
- Optimized for CI environments
- Includes all dependencies for full build capability

## Quick Start

### GitHub Actions
```bash
cp examples/ci/github-actions.yml .github/workflows/loko-validate.yml
git add .github/workflows/loko-validate.yml
git commit -m "Add loko architecture validation"
git push
```

### GitLab CI
```bash
cp examples/ci/.gitlab-ci.yml .gitlab-ci.yml
git add .gitlab-ci.yml
git commit -m "Add loko architecture validation"
git push
```

### Docker Compose (Local Development)
```bash
cd examples/ci
docker-compose up
# Edit your architecture files - docs rebuild automatically
```

## Validation Flags

- `--strict`: Treat warnings as errors (recommended for CI)
- `--exit-code`: Return non-zero exit code on validation failures (required for CI)

## Documentation

See [docs/guides/ci-cd-integration.md](../../docs/guides/ci-cd-integration.md) for detailed setup instructions and troubleshooting.

## Testing

All CI examples are tested in real pipelines before release. See `tests/ci/` for contract tests.
