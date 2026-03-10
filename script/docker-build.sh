#!/bin/bash
set -e

REGISTRY="git.pku.edu.cn/2200011523"
IMAGE_NAME="answer"
VERSION="2.0.0"

echo "Building Apache Answer with MCP Support..."
echo "Registry: ${REGISTRY}/${IMAGE_NAME}"

docker build \
  --build-arg GOPROXY="https://goproxy.cn,direct" \
  --tag "${REGISTRY}/${IMAGE_NAME}:latest" \
  --tag "${REGISTRY}/${IMAGE_NAME}:${VERSION}" \
  .

echo "Pushing images to registry..."
docker push "${REGISTRY}/${IMAGE_NAME}:latest"
docker push "${REGISTRY}/${IMAGE_NAME}:${VERSION}"

echo "Build and push completed successfully!"
echo "Image: ${REGISTRY}/${IMAGE_NAME}:latest"
echo "Version: ${VERSION}"
