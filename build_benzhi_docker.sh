#!/bin/bash
set -e

IMAGE_NAME=${1:-archive-orchestrator}
DOCKER_PLATFORM=${2:-linux/amd64}

docker build --platform "$DOCKER_PLATFORM" -f benzhi.Dockerfile -t "$IMAGE_NAME" .

echo ""
echo "Docker image '$IMAGE_NAME' built successfully!"
echo ""
echo "Run an interactive shell with: docker run -it $IMAGE_NAME:latest"
