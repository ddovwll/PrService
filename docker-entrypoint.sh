#!/usr/bin/env sh
set -e

echo "Waiting for database at ${DB_HOST}:${DB_PORT}..."

while ! nc -z "${DB_HOST:-postgres}" "${DB_PORT:-5432}"; do
  echo "Postgres is unavailable - sleeping"
  sleep 1
done

echo "Postgres is up - running migrations"

/app/migrator

echo "Migrations completed - starting API"

exec /app/http_api