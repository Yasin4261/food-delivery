# Food Delivery — Freelance Home-Chef Platform

A backend API for a freelance home-chef food delivery platform (a basic Uber
Eats clone). Freelance chefs publish menus and dishes; customers discover nearby
chefs, order (possibly from several chefs at once), pay cash or card, and track
their order through to delivery.

Built in Go with a **hexagonal (ports & adapters) architecture** and fully
Dockerized. For the architecture deep-dive, database plan, and the
feature-by-feature build recipe, see [CLAUDE.md](./CLAUDE.md).

> **Status:** actively being rebuilt from scratch on a clean foundation.
> Implemented today: health check + **authentication** (register, login, logout,
> profile). Chefs, menus, orders, favorites, reviews, chat and earnings are on
> the roadmap.

---

## Tech stack

- **Go 1.25**, standard library `net/http` (1.22+ method-based routing)
- **PostgreSQL 16** via `database/sql` + `lib/pq` (raw SQL, no ORM)
- **golang-migrate** for versioned schema migrations
- **JWT** auth (`golang-jwt/v5`) + **bcrypt** password hashing
- **Docker** + Docker Compose (app, database, Adminer DB UI)

## Architecture

Dependencies point inward: `handler → service → domain`. The core (domain +
service) never imports HTTP or SQL; the outside world plugs in through
interfaces (ports).

```
cmd/api/main.go        composition root — wires adapters into the core
config/                environment configuration
database/              connection pool + migration runner
migrations/            versioned SQL (*.up.sql / *.down.sql)
internal/
  domain/              entities + repository interfaces (the core)
  service/             use cases / business logic
  repository/          PostgreSQL adapters (the only place with SQL)
  handler/             HTTP handlers
  middleware/          JWT authentication
  router/              route table
```

## Getting started

### With Docker (recommended)

```bash
docker compose up --build
```

This starts:

| Service | URL | Notes |
|---|---|---|
| API | http://localhost:8080 | runs migrations on startup |
| PostgreSQL | localhost:5432 | `postgres` / `postgres123`, db `food_delivery` |
| Adminer | http://localhost:8081 | web DB UI (server: `db`) |

Verify it is up:

```bash
curl http://localhost:8080/health
# {"database":"ok","status":"ok"}
```

### Locally (without Docker)

Requires Go 1.25+ and a running PostgreSQL. Copy the example env and run:

```bash
cp .env.example .env      # adjust DATABASE_URL if needed
make run                  # go run ./cmd/api
```

## Configuration

Configuration is read from environment variables (a local `.env` file is loaded
if present). See [`.env.example`](./.env.example).

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `ENV` | `development` | environment name |
| `DATABASE_URL` | — | PostgreSQL DSN (**required**) |
| `JWT_SECRET` | — | secret used to sign JWTs |
| `JWT_EXPIRATION` | `24h` | token lifetime (Go duration) |
| `AUTO_MIGRATE` | `false` | run migrations on startup when `true` |

## API

Base path: `/api/v2`. All responses are JSON.

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | — | liveness + DB reachability |
| `POST` | `/api/v2/auth/register` | — | create an account, returns a JWT |
| `POST` | `/api/v2/auth/login` | — | authenticate, returns a JWT |
| `POST` | `/api/v2/auth/logout` | — | client discards its token |
| `GET` | `/api/v2/auth/me` | Bearer | current user's profile |

Example:

```bash
# Register (returns {token, expires_at, user})
curl -X POST http://localhost:8080/api/v2/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"yasin","email":"yasin@example.com","password":"secret123","role":"chef"}'

# Use the token on a protected route
curl http://localhost:8080/api/v2/auth/me \
  -H "Authorization: Bearer <token>"
```

Roles are `customer` (default), `chef`, and `admin`.

## Testing

Business logic is tested without a database: services depend on repository
*interfaces*, so tests inject in-memory fakes.

```bash
make test                          # go test ./...
go test -race -cover ./...         # race detector + coverage
go test -coverprofile=cover.out ./... && go tool cover -html=cover.out
```

What is covered today:

- `internal/domain` — entity defaults and role validation
- `internal/service` — register/login rules, password hashing, JWT round-trip,
  error mapping (table-driven, with a fake repository)
- `internal/handler` — full register → login → `/me` flow over the real router
  and middleware, plus HTTP error codes

## Database & migrations

Schema lives in `migrations/` as ordered `*.up.sql` / `*.down.sql` pairs,
applied by golang-migrate (automatically on startup when `AUTO_MIGRATE=true`).

```bash
make migrate-up                    # apply pending migrations
make migrate-down                  # roll back the last migration
make migrate-create name=create_favorites_table
```

## Common commands

```bash
make run            # run locally
make dev            # docker compose dev stack (hot reload via air)
make build          # build binary to bin/
make test           # run tests
make migrate-up     # apply migrations
make fmt            # go fmt ./...
```

## Roadmap

Following the inside-out recipe in [CLAUDE.md §2](./CLAUDE.md):

- [x] Dockerized hexagonal skeleton + health check
- [x] Authentication (register, login, logout, profile)
- [ ] Forgot / reset password
- [ ] Chef profiles (create, list, nearby by location)
- [ ] Menus & dishes
- [ ] Orders (multi-chef cart, status lifecycle, cash/card)
- [ ] Favorites, reviews & ratings
- [ ] Real-time chat
- [ ] Chef earnings & online/offline status

## License

MIT — see [LICENSE](./LICENSE).
