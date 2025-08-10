#!/bin/sh

# Backend startup script for Docker
# This script runs database migrations before starting the server

set -e

echo "Starting backend initialization..."

# Wait for database to be ready
echo "Waiting for database to be ready..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  echo "Database is unavailable - sleeping"
  sleep 2
done

echo "Database is ready!"

# Run database migrations
echo "Running database migrations..."
cd /app
./migrate -direction up || {
  echo "Migration failed, but continuing with server startup..."
}

echo "Starting server..."
exec ./server
