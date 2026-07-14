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
TMP_FILE="$(mktemp "$BACKUP_DIR/.patchnotes-${TS}.XXXXXX.dump")"

cleanup() {
  if [ -n "${TMP_FILE:-}" ]; then
    rm -f -- "$TMP_FILE"
  fi
}
trap cleanup EXIT

/usr/bin/docker-compose exec -T db \
  pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc > "$TMP_FILE"

if [ ! -s "$TMP_FILE" ]; then
  echo "pg_dump produced an empty backup"
  exit 1
fi

mv -- "$TMP_FILE" "$OUT_FILE"
TMP_FILE=""

find "$BACKUP_DIR" -type f -name 'patchnotes-*.dump' -mtime +14 -delete

echo "Wrote $OUT_FILE"
