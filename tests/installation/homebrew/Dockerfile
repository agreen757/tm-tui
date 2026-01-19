FROM node:20-slim

# Install basic utilities
RUN apt-get update && apt-get install -y \
    curl \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

# Install Go (required for building the project)
RUN curl -OL https://golang.org/dl/go1.23.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz && \
    rm go1.23.0.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/root/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Set working directory
WORKDIR /workspace

# Copy project files
COPY . .

# Set up a clean PATH for testing
# Node.js from official image should be at /usr/local/bin
ENV NPM_PREFIX="/usr/local"

# This Dockerfile simulates Homebrew-style installation:
# - Node.js and npm are installed system-wide
# - npm global packages go to /usr/local/bin (already in PATH)
# - Similar to how Homebrew installs Node.js on macOS

LABEL description="Test environment simulating Homebrew Node.js installation"
