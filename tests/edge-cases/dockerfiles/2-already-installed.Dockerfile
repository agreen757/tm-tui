# Test Case 2: Already installed task-master
# Tests idempotency of installation
FROM node:18-slim

# Install task-master-ai first
RUN npm install -g task-master-ai

# Install other dependencies
RUN apt-get update && apt-get install -y \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

RUN useradd -m -s /bin/bash testuser && \
    chown -R testuser:testuser /workspace

USER testuser

CMD ["/bin/bash"]
