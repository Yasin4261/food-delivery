#!/bin/sh
# Nightly PostgreSQL backup for the food-delivery stack (#74).
#
# Runs inside the db-backup sidecar (postgres:16-alpine, so pg_dump matches
# the server). Connection settings come from the standard PG* environment
# variables — never from arguments, so credentials never show up in `ps` or
# logs.
#
#   PGHOST / PGUSER / PGPASSWORD / PGDATABASE   connection
#   BACKUP_DIR              where dumps land            (default /backups)
#   BACKUP_RETENTION_DAYS   prune dumps older than this (default 14)
#
# Output: <db>_YYYYmmdd_HHMMSS.dump  — pg_dump custom format (compressed,
# restorable with pg_restore, supports parallel + selective restore).
# When /uploads is mounted read-only, photos are tarred alongside the dump.
set -eu

BACKUP_DIR="${BACKUP_DIR:-/backups}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-14}"
STAMP="$(date +%Y%m%d_%H%M%S)"
DB="${PGDATABASE:?PGDATABASE must be set}"

mkdir -p "$BACKUP_DIR"

DUMP="$BACKUP_DIR/${DB}_${STAMP}.dump"
TMP="$DUMP.partial"

echo "backup: dumping $DB to $DUMP"
pg_dump --format=custom --file="$TMP"
mv "$TMP" "$DUMP"
echo "backup: done ($(du -h "$DUMP" | cut -f1))"

# Uploaded photos live outside the database; snapshot them too when mounted.
if [ -d /uploads ]; then
  UP="$BACKUP_DIR/uploads_${STAMP}.tar.gz"
  tar -czf "$UP.partial" -C /uploads . && mv "$UP.partial" "$UP"
  echo "backup: uploads snapshot $(du -h "$UP" | cut -f1)"
fi

# Retention: drop anything older than the window. .partial files from crashed
# runs are junk after a day.
find "$BACKUP_DIR" -maxdepth 1 -name "${DB}_*.dump" -mtime "+$RETENTION_DAYS" -delete
find "$BACKUP_DIR" -maxdepth 1 -name "uploads_*.tar.gz" -mtime "+$RETENTION_DAYS" -delete
find "$BACKUP_DIR" -maxdepth 1 -name "*.partial" -mtime +1 -delete
echo "backup: retention pruned (> ${RETENTION_DAYS} days)"
