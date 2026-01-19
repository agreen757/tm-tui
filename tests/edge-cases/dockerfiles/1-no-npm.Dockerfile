# Test Case 1: Missing npm
# Tests error handling when npm is not installed
FROM debian:bullseye-slim

# Install only essential tools, intentionally skip npm
RUN apt-get update && apt-get install -y \
    curl \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Set up a user to test non-root scenarios
RUN useradd -m -s /bin/bash testuser && \
    chown -R testuser:testuser /workspace

USER testuser

CMD ["/bin/bash"]
