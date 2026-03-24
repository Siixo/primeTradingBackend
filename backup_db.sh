#!/bin/bash
# Description: Automated PostgreSQL backup script for PrimeTrading
# Usage: Run via cron on the Oracle VM (e.g., daily at 3:00 AM)

set -e

# Configuration
BACKUP_DIR="/home/ubuntu/backups" # Adjust if your user isn't 'ubuntu'
DATE=$(date +"%Y-%m-%d_%H-%M-%S")
CONTAINER_NAME="postgres_db"
DB_USER="user"
DB_NAME="primetrading"

echo "Starting database backup process..."

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Run pg_dump inside the docker container and save securely to the host
docker exec -t $CONTAINER_NAME pg_dump -U $DB_USER -d $DB_NAME -F t > "$BACKUP_DIR/db_backup_$DATE.tar"

# Compress the backup
gzip "$BACKUP_DIR/db_backup_$DATE.tar"

# Keep only the last 7 days of backups to prevent the 50GB disk from filling up
find "$BACKUP_DIR" -name "db_backup_*.tar.gz" -type f -mtime +7 -delete

echo "Success: Backup db_backup_$DATE.tar.gz created securely."
