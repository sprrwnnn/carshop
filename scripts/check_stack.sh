#!/usr/bin/env bash
set -euo pipefail

COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-carshop}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yaml}"
REQUIRED_SERVICES="${REQUIRED_SERVICES:-database cache rabbitmq backend notification prometheus grafana alertmanager mailpit}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is not installed or is not in PATH" >&2
  exit 1
fi

echo "Checking Docker Compose stack '${COMPOSE_PROJECT_NAME}' using ${COMPOSE_FILE}"
docker compose -p "${COMPOSE_PROJECT_NAME}" -f "${COMPOSE_FILE}" ps

running_services="$(docker compose -p "${COMPOSE_PROJECT_NAME}" -f "${COMPOSE_FILE}" ps --services --filter status=running)"
failed=0

for service in ${REQUIRED_SERVICES}; do
  if printf '%s\n' "${running_services}" | grep -qx "${service}"; then
    echo "OK: ${service} is running"
  else
    echo "FAIL: ${service} is not running" >&2
    failed=1
  fi
done

exit "${failed}"
