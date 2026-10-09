# CI examples

Run loko in a pipeline so a pull request cannot break the architecture model.

| File | What it does |
|---|---|
| [github-actions.yml](./github-actions.yml) | Validates with `--strict`, checks formatting, builds the site and uploads it as an artifact. A second job does the same with the Docker image. |
| [.gitlab-ci.yml](./.gitlab-ci.yml) | The same checks for GitLab CI. |
| [docker-compose.yml](./docker-compose.yml) | `validate`, `build` and `serve` from `ghcr.io/madstone-tech/loko:latest`, for local use without installing loko. |

## What the checks mean

- `loko validate --strict` exits `0` when clean, `1` on errors and `2` on warnings. Warnings
  include elements without `docs`, elements with no relationships, systems with no containers
  and deployment instances with no binding.
- `loko fmt --check` lists files that are not in canonical form and exits `1`. Run `loko fmt`
  locally to fix them.
- `loko build --out site` renders diagrams (d2 and SVG), markdown and an HTML site. Rendering is
  in-process; no `d2` binary or other tool is needed.

## Installing loko

```bash
brew install --cask madstone-tech/tap/loko             # macOS and Linux
go install github.com/madstone-tech/loko@latest        # any platform with Go
docker run --rm -v "$PWD:/workspace" ghcr.io/madstone-tech/loko:latest validate --strict
```

Pin a release tag instead of `latest` for reproducible pipelines.

## Docker notes

- The image is distroless: `loko` is the entrypoint, `/workspace` is the working directory, and
  there is no shell. Pass loko arguments directly.
- It runs as uid 65532. When `build` writes into a mounted directory, run it as your own user
  (`--user "$(id -u):$(id -g)"`).
- `loko serve` listens on `127.0.0.1` only. In a container it needs host networking
  (see `docker-compose.yml`); otherwise run `loko serve` natively.
