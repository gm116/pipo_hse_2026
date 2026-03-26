#!/bin/sh
set -eu

INTERVAL_SECONDS="${BACKUP_INTERVAL_SECONDS:-86400}"
RUN_ON_START="${BACKUP_RUN_ON_START:-false}"

if [ "${RUN_ON_START}" = "true" ]; then
  /scripts/db_backup.sh
fi

while true; do
  sleep "${INTERVAL_SECONDS}"
  /scripts/db_backup.sh
done
