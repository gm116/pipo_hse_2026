#!/bin/sh
set -eu

: "${POSTGRES_HOST:?POSTGRES_HOST is required}"
: "${POSTGRES_PORT:?POSTGRES_PORT is required}"
: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"

RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"
PREFIX="${BACKUP_FILE_PREFIX:-task_tracker}"

mkdir -p /backups

TIMESTAMP="$(date -u +%Y%m%d_%H%M%S)"
OUT_FILE="/backups/${PREFIX}_${TIMESTAMP}.sql.gz"

export PGPASSWORD="${POSTGRES_PASSWORD}"

echo "[backup] start ${TIMESTAMP}"
pg_dump \
  -h "${POSTGRES_HOST}" \
  -p "${POSTGRES_PORT}" \
  -U "${POSTGRES_USER}" \
  -d "${POSTGRES_DB}" \
  --no-owner \
  --no-privileges \
  | gzip -9 > "${OUT_FILE}"

echo "[backup] done ${OUT_FILE}"

find /backups -type f -name "${PREFIX}_*.sql.gz" -mtime +"${RETENTION_DAYS}" -print -delete
