#!/bin/sh
# db-backup sidecar entrypoint: take one backup at startup (so a fresh deploy
# is covered immediately and misconfiguration surfaces in the logs right
# away), then hand over to cron for the nightly schedule.
set -eu

SCHEDULE="${BACKUP_SCHEDULE:-0 3 * * *}"

/usr/local/bin/backup.sh || echo "backup: initial run failed (cron will retry on schedule)"

echo "$SCHEDULE /usr/local/bin/backup.sh" > /etc/crontabs/root
echo "backup: scheduled '$SCHEDULE'"
exec crond -f -l 8
