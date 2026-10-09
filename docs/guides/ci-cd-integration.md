# CI/CD Integration

Run loko in CI to keep the architecture source valid, formatted, and published. Three commands
cover it:

| Step | Command | Fails when |
|---|---|---|
| Format | `loko fmt --check` | a `*.loko.hcl` file is not in canonical form (exit `1`; the files are listed) |
| Validate | `loko validate --strict` | errors (exit `1`) or warnings (exit `2`) |
| Publish | `loko build --out public` | errors; nothing is written when compilation fails |

Exit codes are the same for every command: `0` clean, `1` errors, `2` warnings under `--strict`.
Drop `--strict` if warnings such as `docs_not_found` should not block a merge.

## Machine-readable diagnostics

`loko validate --format json` prints the diagnostics contract on stdout. The exit code is the same
as with text output.

```json
{
  "schemaVersion": 1,
  "diagnostics": [
    {
      "severity": "error",
      "code": "unresolved_reference",
      "summary": "Unresolvable reference",
      "detail": "...",
      "range": { "file": "bad.loko.hcl", "startLine": 2, "startColumn": 12, "...": "..." }
    }
  ],
  "summary": { "errors": 1, "warnings": 0 }
}
```

Use it to annotate pull requests. For GitHub Actions workflow commands:

```bash
loko validate --format json | jq -r '.diagnostics[] |
  "::\(if .severity == "error" then "error" else "warning" end) file=\(.range.file),line=\(.range.startLine),col=\(.range.startColumn)::\(.code): \(.summary)"'
```

Diagnostics may also carry an `address` and `related` ranges (the first declaration of a
duplicate, for instance).

## Installing loko in CI

- **Container image**: `ghcr.io/madstone-tech/loko:<tag>`, with tags such as `v1.0.0`, `v1.0` and
  `latest`. `loko` is the entrypoint and the working directory is `/workspace`. The image is
  distroless: it has no shell, so run it with `docker run` rather than as a job image whose
  script needs a shell. It runs as uid 65532; pass `--user` so `build` can write to the mounted
  checkout. The examples here pin `v1.0`, which takes patch releases but never a new minor
  version, so a CI run cannot change behaviour under you; use `latest` for trying loko out.

  ```bash
  docker run --rm -v "$PWD:/workspace" ghcr.io/madstone-tech/loko:v1.0 validate --strict
  docker run --rm --user "$(id -u):$(id -g)" -v "$PWD:/workspace" \
    ghcr.io/madstone-tech/loko:v1.0 build --out public
  ```

- **Go toolchain** (Go 1.27 or later): `go install github.com/madstone-tech/loko@v1.0.1`.
- **Release archives**: download from the GitHub releases page and put `loko` on `PATH`.

No other program is needed. Diagrams render in-process; there is no `d2` binary to install.

Pin a version. The language is stable within v1.x, but a newer construct fails on an older tool,
so keep the CI version at least as new as the `loko_version` your project declares.

## GitHub Actions

```yaml
# .github/workflows/architecture.yml
name: Architecture

on:
  pull_request:
    paths: ["**/*.loko.hcl", "docs/**", "templates/**"]
  push:
    branches: [main]

jobs:
  check:
    runs-on: ubuntu-latest
    env:
      LOKO: docker run --rm --user 1001:1001 -v ${{ github.workspace }}:/workspace ghcr.io/madstone-tech/loko:v1.0
    steps:
      - uses: actions/checkout@v7

      - name: Format
        run: $LOKO fmt --check

      - name: Validate
        run: $LOKO validate --strict

      - name: Build site
        if: github.ref == 'refs/heads/main'
        run: $LOKO build --out public

      - uses: actions/upload-pages-artifact@v5
        if: github.ref == 'refs/heads/main'
        with:
          path: public
```

`1001:1001` is the runner user on GitHub-hosted Ubuntu runners; use `$(id -u):$(id -g)` in a
`run` step if yours differs. If the architecture lives in a subdirectory, add `-p architecture`
(or `--project architecture`) to each command. Add a `deploy-pages` job to publish the uploaded
artifact.

## GitLab CI

GitLab runs each job's script inside the job image, and the loko image has no shell, so install
the binary in a Go image instead:

```yaml
# .gitlab-ci.yml
architecture:
  image: golang:1.27
  before_script:
    - go install github.com/madstone-tech/loko@v1.0.1
  script:
    - loko fmt --check
    - loko validate --strict
    - loko validate --format json > loko-diagnostics.json
  artifacts:
    when: always
    paths: [loko-diagnostics.json]
  rules:
    - changes: ["**/*.loko.hcl", "docs/**", "templates/**"]

pages:
  image: golang:1.27
  before_script:
    - go install github.com/madstone-tech/loko@v1.0.1
  script:
    - loko build --out public
  artifacts:
    paths: [public]
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
```

With a Docker-in-Docker runner you can call the GHCR image with `docker run` as in the GitHub
example instead.

## Committing built output

`loko build` output is byte-identical across runs and machines, so you can commit `dist/` and
check that it is current:

```bash
loko build
git diff --exit-code -- dist/
```

Every generated file carries a notice naming its sources, and `build` only removes files listed in
`dist/.loko-manifest`, so other files in the directory are safe.

## Exporting the model

`loko export --format json --out architecture.json` writes the compiled architecture as a
byte-stable artifact for other tools; `--format toon` writes the token-efficient form (see the
[TOON guide](toon-format-guide.md)). Nothing is written when compilation reports errors.

## Troubleshooting

| Symptom | Cause |
|---|---|
| `fmt --check` exits `1` | Run `loko fmt` locally and commit the result |
| `validate --strict` exits `2` | Warnings only; fix them or drop `--strict` |
| `version_unsatisfied` | The CI loko is older than the project's `loko_version` constraint |
| `build` fails with `permission denied` in Docker | The container's uid cannot write the mount; pass `--user` |
| `build` fails with `theme_invalid` | A file in `templates/` is misnamed or malformed; see [Site theming](site-theming.md) |

See the [CLI reference](../cli-reference.md) for every flag.
