# syntax=docker/dockerfile:1.7
# =============================================================================
# examples/ci/Dockerfile — Full-featured CI image
#
# Differs from the root Dockerfile (distroless, source-built) in that:
#   - Includes bash, curl, git, and inotify-tools for CI scripts and watch loops
#   - Ubuntu 22.04 LTS base for broad CI runner compatibility
#   - Suitable for CI pipelines and local development via docker compose
#
# Build from repo root:
#   docker build -t loko:ci -f examples/ci/Dockerfile \
#     [--build-arg VERSION=v0.1.0] \
#     [--build-arg D2_VERSION=v0.7.1] \
#     .
#
# Run:
#   docker run --rm -v $(pwd):/workspace loko:ci validate --strict --exit-code
# =============================================================================

# =============================================================================
# Stage 1: loko builder
# =============================================================================
FROM golang:1.25-alpine AS loko-builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

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
# Stage 2: D2 builder
# Builds D2 from source with CGO_ENABLED=0 to produce a portable static binary.
# =============================================================================
FROM golang:1.25-alpine AS d2-builder

RUN apk add --no-cache git ca-certificates

ARG D2_VERSION=v0.7.1

RUN git clone --depth 1 --branch ${D2_VERSION} \
      https://github.com/terrastruct/d2.git /d2

WORKDIR /d2

RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags "-s -w -X oss.terrastruct.com/d2/lib/version.Version=${D2_VERSION}" \
      -o /out/d2 \
      .

# =============================================================================
# Stage 3: runtime
# Ubuntu 22.04 LTS — provides bash, apt, and inotify-tools for CI scripts.
# =============================================================================
FROM ubuntu:22.04

LABEL org.opencontainers.image.title="loko-ci" \
      org.opencontainers.image.description="loko CI image with D2 and shell utilities" \
      org.opencontainers.image.source="https://github.com/madstone-tech/loko" \
      org.opencontainers.image.licenses="BUSL-1.1"

ENV DEBIAN_FRONTEND=noninteractive

# System packages:
#   ca-certificates  – TLS verification
#   curl             – downloading release binaries in CI scripts
#   git              – version detection, changelog tooling
#   inotify-tools    – available for custom file-watch scripts in CI
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        inotify-tools \
    && rm -rf /var/lib/apt/lists/*

COPY --from=loko-builder /out/loko /usr/local/bin/loko
COPY --from=d2-builder   /out/d2   /usr/local/bin/d2

# Scaffold templates for `loko new`.
COPY --from=loko-builder /src/templates /usr/local/share/loko/templates

ENV LOKO_TEMPLATE_DIR=/usr/local/share/loko/templates

# Note: veve-cli (PDF generation) is a private Terrastruct binary and is not
# publicly available. PDF output is disabled in this image.
# If you have access to veve-cli, add it here:
#   COPY veve-cli /usr/local/bin/veve-cli
#   RUN chmod +x /usr/local/bin/veve-cli

WORKDIR /workspace

RUN loko version && d2 --version

CMD ["loko", "--help"]

# =============================================================================
# Usage
# =============================================================================
#
# Validate (non-zero exit on violation):
#   docker run --rm -v $(pwd):/workspace loko:ci validate --strict --exit-code
#
# Build HTML docs:
#   docker run --rm -v $(pwd):/workspace loko:ci build --format html
#
# Build Markdown docs:
#   docker run --rm -v $(pwd):/workspace loko:ci build --format markdown
#
# Interactive shell:
#   docker run --rm -it -v $(pwd):/workspace loko:ci bash
#
# Expected image size: ~120MB (Ubuntu 22.04 base + utilities + loko + D2)
