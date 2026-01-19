# Test Case 5: Old npm version
# Tests compatibility with older npm versions
FROM node:14-slim

# Node 14 comes with older npm version
RUN apt-get update && apt-get install -y \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

RUN useradd -m -s /bin/bash testuser && \
    chown -R testuser:testuser /workspace

USER testuser

CMD ["/bin/bash"]
