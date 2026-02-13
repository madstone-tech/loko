# Dockerfile for loko with full build capabilities
#
# This image includes:
# - loko binary
# - D2 diagram renderer
# - veve-cli for PDF generation
# - All dependencies for full architecture documentation builds
#
# Build: docker build -t loko:latest -f examples/ci/Dockerfile .
# Run: docker run --rm -v $(pwd):/workspace loko:latest validate --strict

FROM ubuntu:22.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install loko
ARG LOKO_VERSION=latest
RUN curl -L https://github.com/madstone-tech/loko/releases/latest/download/loko-linux-amd64 -o /usr/local/bin/loko \
    && chmod +x /usr/local/bin/loko

# Install D2 diagram renderer
RUN curl -fsSL https://d2lang.com/install.sh | sh -s --

# Install veve-cli for PDF generation
# Note: veve-cli is optional - loko will gracefully degrade if not present
RUN curl -L https://github.com/terrastruct/veve/releases/latest/download/veve-linux-amd64 -o /usr/local/bin/veve-cli \
    && chmod +x /usr/local/bin/veve-cli

# Set working directory
WORKDIR /workspace

# Verify installations
RUN loko version && \
    d2 --version && \
    veve-cli --version || echo "veve-cli optional dependency"

# Default command: show help
CMD ["loko", "--help"]

# Usage examples:
#
# Validate architecture:
#   docker run --rm -v $(pwd):/workspace loko:latest validate --strict --exit-code
#
# Build HTML documentation:
#   docker run --rm -v $(pwd):/workspace loko:latest build --format html
#
# Build PDF documentation:
#   docker run --rm -v $(pwd):/workspace loko:latest build --format pdf
#
# Interactive shell:
#   docker run --rm -it -v $(pwd):/workspace loko:latest /bin/bash
#
# Expected image size: ~250MB (Ubuntu base + loko + D2 + veve-cli)
