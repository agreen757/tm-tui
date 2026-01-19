# Test Case 4: Permission errors
# Tests handling of permission denied scenarios
FROM node:18-slim

RUN apt-get update && apt-get install -y \
    git \
    make \
    sudo \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Create user without sudo access
RUN useradd -m -s /bin/bash testuser && \
    chown -R testuser:testuser /workspace

# Make npm directories owned by root to trigger permission errors
RUN chown -R root:root /usr/local/lib/node_modules && \
    chmod -R 755 /usr/local/lib/node_modules

USER testuser

CMD ["/bin/bash"]
