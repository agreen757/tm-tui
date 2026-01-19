# Docker Test Environment - Quick Start

This guide helps you quickly get started with the Docker test environment setup scripts.

## Prerequisites

- Docker Desktop (macOS/Windows) or Docker Engine (Linux)
- At least 4GB free disk space
- Internet connection for downloading images

## Quick Setup

### 1. Validate Setup (Recommended First Step)

```bash
cd tests/docker
./validate-setup.sh
```

This runs 27 validation tests to ensure everything is configured correctly.

### 2. Setup All Platforms

```bash
cd tests/docker
./setup-all.sh
```

This will build Docker images for:
- Ubuntu 22.04 (apt and nvm)
- Alpine Linux (apk and nvm)
- Windows/Git Bash simulation

### 3. Run Tests

```bash
# Run all platform tests
docker-compose up

# Run specific platforms
docker-compose up ubuntu-apt ubuntu-nvm

# Run individual test
docker run --rm -v "$(pwd)/../..:/workspace:ro" tm-test-ubuntu-apt:latest
```

## Platform-Specific Setup

### Just Ubuntu
```bash
./setup-ubuntu.sh
docker-compose up ubuntu-apt ubuntu-nvm
```

### Just Alpine
```bash
./setup-alpine.sh
docker-compose up alpine-apk alpine-nvm
```

### macOS Testing
```bash
./setup-macos.sh
# Follow on-screen instructions
```

### Windows Testing
```bash
./setup-windows.sh
# Creates Git Bash simulation (Linux-based)
# Or Windows Server container if Windows containers available
```

## Common Commands

### Check Docker
```bash
# Verify Docker is running
docker ps

# Check Docker version
docker --version
```

### Cleanup
```bash
# Stop all test containers
docker-compose down

# Remove test images
docker rmi $(docker images -q 'tm-test-*')

# Full cleanup
docker system prune -a
```

### Debugging
```bash
# View container logs
docker logs <container_name>

# Run container interactively
docker run --rm -it -v "$(pwd)/../..:/workspace" tm-test-ubuntu-apt:latest bash

# Check running containers
docker ps -a
```

## Troubleshooting

### Docker Not Running
```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker
```

### Permission Issues (Linux)
```bash
sudo usermod -aG docker $USER
newgrp docker
```

### Out of Space
```bash
docker system prune -a --volumes
```

## For More Information

See [README.md](README.md) for comprehensive documentation including:
- Detailed platform-specific guides
- Helper function reference
- CI/CD integration examples
- Advanced usage patterns
- Complete troubleshooting guide

## Support

If you encounter issues:
1. Run `./validate-setup.sh` to check configuration
2. Check Docker daemon is running
3. Review logs in `docker logs <container>`
4. See troubleshooting section in README.md
