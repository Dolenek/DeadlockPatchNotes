#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [ ! -f .env ]; then
  echo "Missing .env in $ROOT_DIR"
  exit 1
fi

set -a
source .env
set +a

BACKUP_DIR="$ROOT_DIR/backups"
mkdir -p "$BACKUP_DIR"

TS="$(date -u +%Y%m%d-%H%M%S)"
OUT_FILE="$BACKUP_DIR/patchnotes-${TS}.dump"

/usr/bin/docker-compose exec -T db \
  pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc > "$OUT_FILE"

find "$BACKUP_DIR" -type f -name 'patchnotes-*.dump' -mtime +14 -delete

echo "Wrote $OUT_FILE"
