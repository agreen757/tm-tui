# Test Case 6: Limited disk space
# Tests behavior with disk space constraints
# Note: Actual disk limit is set via docker run --storage-opt
FROM node:18-slim

RUN apt-get update && apt-get install -y \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

RUN useradd -m -s /bin/bash testuser && \
    chown -R testuser:testuser /workspace

USER testuser

CMD ["/bin/bash"]
