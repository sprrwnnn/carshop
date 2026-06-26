#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 path/to/backup.dump" >&2
  exit 1
fi

COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-carshop}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yaml}"
DATABASE_SERVICE="${DATABASE_SERVICE:-database}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-carshop}"
BACKUP_FILE="$1"

if [ ! -f "${BACKUP_FILE}" ]; then
  echo "Backup file does not exist: ${BACKUP_FILE}" >&2
  exit 1
fi

echo "Restoring ${BACKUP_FILE} into ${POSTGRES_DB}"
docker compose -p "${COMPOSE_PROJECT_NAME}" -f "${COMPOSE_FILE}" exec -T "${DATABASE_SERVICE}" \
  pg_restore -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" --clean --if-exists --no-owner --no-acl < "${BACKUP_FILE}"

echo "Restore completed"
