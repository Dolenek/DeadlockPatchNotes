#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SYNC_LOG="/var/log/deadlockpatchnotes-sync.log"
BACKUP_LOG="/var/log/deadlockpatchnotes-backup.log"

TMP_CRON="$(mktemp)"
trap 'rm -f "$TMP_CRON"' EXIT

crontab -l 2>/dev/null | grep -v 'deadlockpatchnotes-' > "$TMP_CRON" || true

echo "0 */6 * * * cd $ROOT_DIR && /bin/bash $ROOT_DIR/scripts/server/run_sync.sh >> $SYNC_LOG 2>&1 # deadlockpatchnotes-sync" >> "$TMP_CRON"
echo "30 3 * * * cd $ROOT_DIR && /bin/bash $ROOT_DIR/scripts/server/backup_postgres.sh >> $BACKUP_LOG 2>&1 # deadlockpatchnotes-backup" >> "$TMP_CRON"

crontab "$TMP_CRON"
echo "Installed cron jobs for sync and backup"
