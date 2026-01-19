# Test Case 3: Misconfigured PATH
# Tests recovery when npm bin directory not in PATH
FROM node:18-slim

RUN apt-get update && apt-get install -y \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

RUN useradd -m -s /bin/bash testuser && \
    chown -R testuser:testuser /workspace

USER testuser

# Set a minimal PATH that excludes npm global bin
ENV PATH=/usr/local/bin:/usr/bin:/bin

CMD ["/bin/bash"]
