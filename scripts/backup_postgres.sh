#!/usr/bin/env bash
set -euo pipefail

COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-carshop}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yaml}"
DATABASE_SERVICE="${DATABASE_SERVICE:-database}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-carshop}"
BACKUP_DIR="${BACKUP_DIR:-backups/postgres}"

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_file="${BACKUP_DIR}/${POSTGRES_DB}_${timestamp}.dump"

mkdir -p "${BACKUP_DIR}"

echo "Creating PostgreSQL backup: ${backup_file}"
docker compose -p "${COMPOSE_PROJECT_NAME}" -f "${COMPOSE_FILE}" exec -T "${DATABASE_SERVICE}" \
  pg_dump -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" --format=custom --no-owner --no-acl > "${backup_file}"

echo "Backup completed: ${backup_file}"
