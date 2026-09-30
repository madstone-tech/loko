# syntax=docker/dockerfile:1.7
# =============================================================================
# Stage 1: loko builder
# Compiles loko with CGO_ENABLED=0 → fully static binary, no libc deps.
# =============================================================================
FROM golang:1.25-alpine AS loko-builder

# git   – required by go mod download for VCS-backed modules
# ca-certificates – HTTPS for module proxy
RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Resolve dependencies first; these layers are cached across source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build arguments mirror .goreleaser.yaml ldflags.
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags "-s -w \
        -X main.version=${VERSION} \
        -X main.commit=${COMMIT} \
        -X main.date=${BUILD_DATE} \
        -X main.builtBy=docker" \
      -o /out/loko .

# =============================================================================
# Stage 2: runtime
# Google Distroless static — no shell, no package manager, no libc.
# ~2 MiB base, Apache-2.0, signed with keyless cosign, daily CVE patches.
#
# Use the :debug tag locally for investigation:
#   FROM gcr.io/distroless/static-debian12:debug
# which adds a busybox shell without changing any other layer.
# =============================================================================
FROM gcr.io/distroless/static-debian12:nonroot

# OCI image annotations (replaces LABEL for spec compliance).
LABEL org.opencontainers.image.title="loko" \
      org.opencontainers.image.description="C4 architecture documentation tool" \
      org.opencontainers.image.source="https://github.com/madstone-tech/loko" \
      org.opencontainers.image.licenses="BUSL-1.1"

# distroless/static already includes ca-certificates.
# The :nonroot tag runs as uid=65532 (nonroot) by default — no USER instruction needed.

# Statically linked loko binary.
COPY --from=loko-builder /out/loko /usr/local/bin/loko

# Nothing else is needed: diagrams render in-process (feature 014, FR-014),
# so the image carries no d2 binary, no shell, and no other executable. A
# successful `loko build` inside it is the proof for US6/AC2.

# Project directories are mounted here at `docker run` time.
WORKDIR /workspace

# All args must be in exec (vector) form — distroless has no shell to exec string form.
# loko is the entrypoint, so `docker run loko:dev build` runs `loko build`.
ENTRYPOINT ["/usr/local/bin/loko"]
CMD ["--help"]

# =============================================================================
# Usage
# =============================================================================
#
# Build (development):
#   docker build -t loko:dev .
#
# Build (production — embed version/commit):
#   docker build \
#     --build-arg VERSION=0.1.0 \
#     --build-arg COMMIT=$(git rev-parse --short HEAD) \
#     --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
#     -t loko:0.1.0 .
#
# Validate architecture (exits non-zero on violations):
#   docker run --rm -v $(pwd):/workspace loko:dev validate --strict
#
# Build diagrams, markdown and the site into ./dist:
#   docker run --rm -v $(pwd):/workspace loko:dev build
#
# Build only the site (svg is added because the site embeds it):
#   docker run --rm -v $(pwd):/workspace loko:dev build --format html
#
# Debug shell (requires the :debug tag in the FROM above):
#   docker run --rm -it --entrypoint /busybox/sh loko:dev
#
# Multi-arch build (requires buildx):
#   docker buildx build --platform linux/amd64,linux/arm64 -t loko:dev .
