# Deploying Home Chef

The production stack is three containers behind **one public origin**:

```
                    ┌─────────────────────────────────────────┐
 browser ──────────►│ web  (Caddy)                            │
  HTTPS (auto-LE)   │  • serves the built Vue SPA (static)    │
                    │  • proxies /api/*, /health, /version ───┼──► api (Go) ──► db (Postgres)
                    └─────────────────────────────────────────┘      internal-only network
```

One origin means **no CORS in production**, the reset-password emails and
iyzico payment callbacks land on the same public URL, and WebSocket chat is
proxied transparently.

## Prerequisites

- A Linux host with Docker + Compose v2
- A domain with an A/AAAA record pointing at the host (for automatic HTTPS)
- SMTP credentials (transactional mail) and iyzico API credentials

## First deployment

```bash
git clone git@github.com:Yasin4261/food-delivery.git && cd food-delivery
cp .env.prod.example .env.prod
$EDITOR .env.prod       # fill EVERY required value (see the template comments)
make prod               # builds SPA + API (git-version stamped) and starts the stack
```

Required values are enforced twice: compose refuses to start on empty
`${VAR:?}` values, and the API **fails fast in ENV=production** on a missing/
placeholder `JWT_SECRET`, missing SMTP or missing iyzico credentials.

For real HTTPS set in `.env.prod`:

```bash
SITE_ADDRESS=food.example.com          # Caddy provisions Let's Encrypt automatically
APP_BASE_URL=https://food.example.com  # public URL used in emails + payment callbacks
```

`SITE_ADDRESS=:80` keeps plain HTTP — only for a local prod test or when TLS
terminates upstream (cloud load balancer).

## Verify

```bash
curl -fsS https://food.example.com/health    # {"database":"ok","status":"ok"}
curl -fsS https://food.example.com/version   # {"version":"v3.3.0","go":"go1.25…"}
# open the domain in a browser: the SPA login page should load
```

## Operations

| Task | Command |
|---|---|
| Update to latest | `git pull && make prod` (rebuilds + restarts; DB volume persists) |
| Logs | `make prod-logs` |
| Stop | `make prod-down` (volumes are kept) |
| Manual backup now | `docker exec food_delivery_backup /usr/local/bin/backup.sh` |
| Backup logs | `docker logs food_delivery_backup` |

## Backups (automatic)

The `db-backup` sidecar takes a **nightly `pg_dump`** (custom format,
compressed) plus a tarball of the uploaded photos, into **`./backups/` on the
host** — they survive `docker compose down -v`. It also runs once at startup,
so a fresh deploy is covered immediately and a misconfiguration shows up in
the logs right away.

Tunables in `.env.prod` (defaults shown):

```bash
BACKUP_SCHEDULE="0 3 * * *"   # cron, container time (UTC)
BACKUP_RETENTION_DAYS=14      # older dumps are pruned automatically
```

**Copy `./backups/` off-site** (rsync/rclone to another machine or object
storage) — a backup on the same disk as the database only protects against
`DROP TABLE`, not against losing the server.

### Restore runbook

1. Stop the API so nothing writes mid-restore: `docker compose -f docker-compose.prod.yml stop api`
2. Pick a dump: `ls -lt backups/ | head`
3. Restore (`--clean --if-exists` drops and recreates objects; add
   `--create -d postgres` instead to rebuild the whole database):

   ```bash
   docker exec -i food_delivery_db_prod \
     pg_restore -U postgres -d food_delivery --clean --if-exists --no-owner \
     < backups/food_delivery_YYYYmmdd_HHMMSS.dump
   ```

4. Photos, if needed: `docker run --rm -v food-delivery_uploads_data:/uploads -v "$PWD/backups:/b" alpine sh -c "tar -xzf /b/uploads_YYYYmmdd_HHMMSS.tar.gz -C /uploads"`
5. Start the API again and smoke-check: `docker compose -f docker-compose.prod.yml start api && curl -fsS localhost/health`

**Drill restores regularly** — restore the latest dump into a scratch
database and compare row counts (see issue #74 for the recorded drill):

```bash
docker exec food_delivery_db_prod createdb -U postgres drill
docker exec -i food_delivery_db_prod pg_restore -U postgres -d drill --no-owner < backups/<latest>.dump
docker exec food_delivery_db_prod psql -U postgres -d drill -c "select count(*) from orders"
docker exec food_delivery_db_prod dropdb -U postgres drill
```

## Notes & limits

- The API runs read-only with `no-new-privileges`; migrations run
  automatically at startup (`AUTO_MIGRATE=true`).
- The token denylist and rate limiter are in-memory by default — correct for
  this **single-instance** stack. To run multiple API instances, add a Redis
  and set `REDIS_URL` (e.g. `redis://redis:6379/0`): revocation and rate
  limits are then shared across instances. On Redis errors both fail **open**
  (availability over strictness) with logged warnings.
- iyzico starts against the **sandbox** (`IYZICO_BASE_URL`); switch to
  `https://api.iyzipay.com` after verifying with test cards (issue #51).
- Postgres is reachable **only** on the internal Docker network; the backup
  sidecar reaches it there too — 5432 is never exposed to the host.
