#!/bin/bash

echo "Starting PostgreSQL database..."
docker-compose up -d postgres

echo "Waiting for database to be ready..."
until docker exec steam-observer-db pg_isready -U steam_user -d steam_observer; do
  sleep 1
done

echo "Database is ready!"
