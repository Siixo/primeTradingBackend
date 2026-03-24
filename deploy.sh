#!/bin/bash
set -e

echo "🚀 Starting Production Deployment..."

# 1. Pull latest code from main
echo "📥 Pulling latest changes from origin main..."
git pull origin main

# 2. Update dependencies and vendor
echo "📦 Updating dependencies..."
go mod tidy
go mod vendor

# 3. Build and Restart Containers
echo "🏗️ Rebuilding and restarting containers..."
# NOTE: We use "docker compose" (with a space) to avoid the KeyError: 'ContainerConfig'
# which is a known bug in the old docker-compose v1.
docker compose up -d --build

# 4. Cleanup
echo "🧹 Cleaning up old images..."
docker image prune -f

echo "✅ Deployment Successful!"
