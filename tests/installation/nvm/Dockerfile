FROM ubuntu:22.04

# Install basic utilities
RUN apt-get update && apt-get install -y \
    curl \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

# Install nvm (Node Version Manager)
ENV NVM_DIR="/root/.nvm"
ENV NODE_VERSION="20.11.0"

RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash && \
    . "$NVM_DIR/nvm.sh" && \
    nvm install $NODE_VERSION && \
    nvm use $NODE_VERSION && \
    nvm alias default $NODE_VERSION

# Configure PATH for nvm
ENV PATH="${NVM_DIR}/versions/node/v${NODE_VERSION}/bin:${PATH}"

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

# This Dockerfile simulates nvm installation:
# - Node.js is installed per-user via nvm
# - npm global packages go to $NVM_DIR/versions/node/v{version}/bin
# - PATH must be configured to include nvm's bin directory

LABEL description="Test environment simulating nvm Node.js installation"
