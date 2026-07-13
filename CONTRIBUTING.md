# Contributing

Thanks for helping build this freelance home-chef delivery platform. This document is the short version; [CLAUDE.md](CLAUDE.md) is the long one (architecture §2, conventions §7, testing §7a, security §10, versioning §9) — when they disagree, CLAUDE.md wins.

## Getting started

```bash
git clone git@github.com:Yasin4261/food-delivery.git && cd food-delivery
cp .env.example .env               # dev defaults work out of the box
docker compose up --build          # Postgres + API (:8080) + Adminer (:8081)
cd web && npm install && npm run dev   # SPA on :5173, /api proxied to the Go API
```

`make dev` runs the hot-reload stack (air). `GET /health` should return `{"status":"ok","database":"ok"}`.

## Architecture in one paragraph

Hexagonal: `domain/` (entities + port interfaces, no framework imports) ← `service/` (use cases) ← `repository/` (Postgres adapters, the **only** place SQL lives) and `handler/`+`router/` (HTTP, no business rules). `cmd/api/main.go` wires everything. **Dependencies point inward — always.** New features are built inside-out: domain → port → service → repository → migration → handler → wiring. The full recipe with an example is in CLAUDE.md §2.

## Workflow

1. Pick or open an issue; discuss design **in the issue first** for schema/lifecycle changes.
2. Branch from `main`: `feature/<slug>`, `fix/<slug>`, `docs/<slug>`, `ops/<slug>`.
3. Implement **with tests** (see below). A feature without tests is not done.
4. `make fmt && go vet ./... && go test -race ./...` must pass locally.
5. Open a PR that says what changed and how it was verified; link the issue (`Closes #N`).
6. All five CI jobs must be green (build+test, Postgres/Redis integration, web build, Playwright E2E, Docker build). Merge, delete the branch.

## Tests (the matrix)

| Layer | Where | Command |
|---|---|---|
| Domain / service / handler (fakes, no DB) | `internal/**/*_test.go` | `make test` |
| Repository + migrations (real Postgres) | `internal/repository/*_integration_test.go` (`//go:build integration`) | `make test-integration` |
| SPA units (stores, router, components) | `web/src/**/*.test.js` (Vitest) | `cd web && npm test` |
| Browser golden path | `web/e2e/*.spec.js` (Playwright) | `cd web && npm run test:e2e` (stack must be up) |

Patterns to copy: `auth_service_test.go` (table-driven service tests over a fake repo), `auth_handler_test.go` (httptest through the real router), `order_repository_integration_test.go` (real SQL). Every new endpoint ships tests for its 401/403 paths — that's policy (§10.5).

## Conventions that will come up in review

- SQL only in `repository/`, only `$1…` placeholders. Business rules only in `domain/`+`service/`. HTTP only in `handler/`.
- Mutate entities through their methods (`order.MarkReady()`), never by assigning fields.
- Nullable columns → pointer fields; wrap errors with `fmt.Errorf("...: %w", err)`.
- One numbered migration pair (`.up.sql`/`.down.sql`) per schema change: `make migrate-create name=...`.
- UI strings go through vue-i18n — **update both** `web/src/i18n/en.js` and `tr.js` (literal `@` must be written `{'@'}`). `en.js` wording is load-bearing for E2E selectors.
- Never commit secrets; new required config must fail fast outside development.
- Security-sensitive branches (auth, payments, file handling, new deps) follow CLAUDE.md §10.

## Documentation is part of the change

Docs are updated **in the same PR** as the code they describe:

- `README.md` / `web/README.md` — user-visible features, setup, commands
- `CLAUDE.md` — §6 current state per feature, §9 tag history per release, §10 when the security posture moves
- `SECURITY.md` — posture summary when auth/payment behaviour changes
- `DEPLOY.md` — anything operational

A PR that changes behaviour but leaves the docs stale is incomplete.

## Releases

SemVer via annotated git tags on green `main` (`vMAJOR.MINOR.PATCH`) + a GitHub release; no long-lived version branches. Details and bump rules in CLAUDE.md §9.
