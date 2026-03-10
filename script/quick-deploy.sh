#!/bin/bash
set -e

echo "Apache Answer Quick Deploy Script"
echo "=================================="

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "Error: Docker Compose is not installed"
    exit 1
fi

# Pull latest image
echo "Pulling latest image..."
docker pull git.pku.edu.cn/2200011523/answer:latest

# Start services
echo "Starting services..."
docker-compose -f docker-compose.prod.yaml up -d

# Wait for service to be ready
echo "Waiting for service to start..."
sleep 5

# Show status
echo ""
echo "Deployment complete!"
echo "===================="
echo "Web UI: http://localhost:9080"
echo "Admin Panel: http://localhost:9080/admin"
echo ""
echo "Check logs with: docker-compose -f docker-compose.prod.yaml logs -f"
echo "Stop services with: docker-compose -f docker-compose.prod.yaml down"
