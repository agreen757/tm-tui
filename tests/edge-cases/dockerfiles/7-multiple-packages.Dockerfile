# Test Case 7: Multiple global packages
# Tests isolation with many existing npm packages
FROM node:18-slim

# Install many common global packages
RUN npm install -g \
    typescript \
    eslint \
    prettier \
    webpack \
    webpack-cli \
    @angular/cli \
    create-react-app \
    vue-cli

RUN apt-get update && apt-get install -y \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

RUN useradd -m -s /bin/bash testuser && \
    chown -R testuser:testuser /workspace

USER testuser

CMD ["/bin/bash"]
