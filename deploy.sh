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

# 3. Force Cleanup and Rebuild
echo "🏗️  Removing existing containers and rebuilding..."
# This step helps avoid the 'ContainerConfig' error in old docker-compose v1
# by ensuring we start with a clean state.
docker-compose down || true
docker rm -f go_app postgres_db || true
docker-compose up -d --build

# 4. Cleanup
echo "🧹 Cleaning up old images..."
docker image prune -f

echo "✅ Deployment Successful!"
