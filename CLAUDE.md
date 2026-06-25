# CLAUDE.md

Guidance for working in this repository. This is a **freelance home-chef food delivery platform** (a basic Uber Eats clone): freelance chefs publish menus/products from a chef app, and customers order from a user app.

---

## 1. What we are building (product scope)

Two kinds of end users sit on top of one backend API:

- **Customer (user app):** browse/search chefs and food, order (possibly from **multiple chefs in one cart**), pay **cash or card**, track order status, see order history, favorite chefs, rate chefs and products, chat with the chef.
- **Chef (chef app):** a freelancer. Create/edit menus and products, set themselves **online/offline**, **accept or decline** incoming orders, update delivery status, see **earnings**, manage kitchen profile (name, kitchen type, phone, address, coordinates).

Shared/common features: **register, login, logout, forgot password**.

Key product rules that shape the data model:
- A chef has a **location (lat/lng)** and a **delivery radius**. A customer only sees chefs that can deliver to the customer's location.
- A single customer order can contain items from **several chefs** (multi-chef cart). Each item remembers which chef it came from.
- Orders move through a **status lifecycle** (see §4). Payment is **cash or card**.
- Customers and chefs can **chat in real time** or just call each other by phone.

> The code was rebuilt from scratch into a clean, Dockerized skeleton (only a `/health` endpoint so far). Everything above beyond health is **planned but not yet built** — see §6 for current state and §5 for the DB tables to add.

---

## 2. Hexagonal architecture — explained for a beginner

You said you're new to hexagonal architecture, so here is the mental model **and** exactly how it maps to the folders in this repo.

### The idea in one picture

```
                  ┌──────────────────────────────────────┐
   HTTP request   │            APPLICATION CORE           │   PostgreSQL
   ───────────►   │                                       │  ◄──────────►
   (driving side) │   domain  ──ports──►  services        │  (driven side)
   handler/       │   (entities + rules + interfaces)     │   repository/
                  └──────────────────────────────────────┘
```

The **core** (domain + services) is the valuable part: your business rules. It must **not** know whether data comes from HTTP, gRPC, or a CLI, nor whether it's stored in Postgres, MySQL, or memory. The outside world talks to the core only through **ports** (Go interfaces). Concrete code that plugs into a port is an **adapter**.

- **Port** = a Go `interface` (a contract). Example: `domain.ChefRepository`.
- **Adapter** = a concrete implementation of that contract. Example: `repository.ChefRepository` (the Postgres version).

The golden rule: **dependencies point inward.** `handler → service → domain`. The domain imports nothing from `handler`, `repository`, or any framework. The repository imports `domain` (to satisfy its interface), never the other way around.

### How this maps to our folders

| Hexagonal concept | This repo | Role |
|---|---|---|
| **Domain / entities** | `internal/domain/*.go` (`user.go`, `chef.go`, `order.go`, `menu.go`, `menu_item.go`) | Pure Go structs + business methods (e.g. `Order.Confirm()`, `Chef.CanDeliver()`). No SQL, no HTTP, no framework imports. |
| **Ports (driven)** | `internal/domain/*_repository.go` | The **interfaces** the core needs from storage, e.g. `ChefRepository`. Defined in `domain` so the core owns the contract. |
| **Driven adapters** | `internal/repository/*.go` | Postgres implementations of those interfaces. The only place raw SQL lives. |
| **Application services** | `internal/service/*.go` | Use cases / orchestration. Takes a request, applies rules via domain methods, calls repositories through their interfaces. **This is where the "what happens when a customer orders" logic lives.** |
| **Driving adapters** | `internal/handler/*.go` + `internal/middleware/*.go` | HTTP. Parse the request → call a service → write JSON. No business rules here. |
| **Composition root** | `cmd/api/main.go` (`initializeApp`) | Wires concrete adapters into the core (dependency injection). The one place that knows about every layer. |
| **Routing** | `internal/router/router.go` | Maps URL+method to a handler method. |

### The "add a feature" recipe (follow this every time)

When you add a feature (say, "customer favorites a chef"), go **inside-out**:

1. **Domain** — add/extend the entity in `internal/domain/`. Add a `Favorite` struct and any rule methods. No DB, no JSON framework logic.
2. **Port** — add the interface method(s) the use case needs, e.g. `FavoriteRepository.Add(...)` in `internal/domain/favorite_repository.go`.
3. **Service** — write the use case in `internal/service/favorite_service.go`: validation + calling domain methods + calling the repository interface.
4. **Adapter (repo)** — implement the interface with SQL in `internal/repository/favorite_repository.go`.
5. **Migration** — add `migrations/0000NN_create_favorites_table.up.sql` and `.down.sql`.
6. **Handler** — parse/return HTTP in `internal/handler/favorite_handler.go`. Keep it thin.
7. **Wire it** — construct repo → service → handler in `cmd/api/main.go` and add routes in `internal/router/router.go`.

If you ever feel tempted to write SQL in a service, or business `if` rules in a handler, stop — that's the layering breaking down.

### Why bother (the payoff)

- You can unit-test a service by passing a **fake repository** (a struct implementing the interface) — no database needed.
- Swapping Postgres for something else means writing one new adapter, touching zero business logic.
- New developers read `service/` to learn *what the app does* without drowning in SQL.

---

## 3. Project layout

```
cmd/api/main.go            # entrypoint + dependency injection (composition root)
config/config.go           # env/config loading
database/database.go       # DB connection pool
database/migrate.go        # golang-migrate runner
migrations/                # versioned SQL (golang-migrate, *.up.sql / *.down.sql)
internal/
  domain/                  # entities + repository interfaces (the core; no external deps)
  service/                 # use cases / business orchestration
  repository/              # Postgres adapters implementing domain interfaces
  handler/                 # HTTP handlers (health_handler.go, response.go)
  router/                  # route table
```

Each empty layer (`domain/`, `service/`, `repository/`) currently holds only a
`doc.go` that states the package's hexagonal contract. Add real files as
features land — `middleware/` (JWT auth) returns when auth is built.

Tech: **Go 1.25**, standard library `net/http` (Go 1.22+ method-based routing like `"POST /api/v2/..."`), `database/sql` + `lib/pq` (raw SQL, no ORM), JWT (`golang-jwt/v5`), bcrypt, `golang-migrate`, Swagger via `swaggo`. Module path: `github.com/Yasin4261/food-delivery`.

---

## 4. Order status lifecycle (target design)

This is the intended `Order` model — to be implemented in `internal/domain/order.go`. Transitions should be enforced by methods that return `ErrInvalidStatusTransition` on an illegal move — **always go through these methods, never set `o.Status` directly.**

```
pending → confirmed → preparing → ready → delivering → delivered
   │           │
   └───────────┴────────► cancelled   (only from pending/confirmed)
```

- `Confirm()` / `StartPreparing()` / `MarkReady()` / `StartDelivering()` / `MarkDelivered()` / `Cancel()`.
- Payment status is separate: `pending → paid | failed | refunded`. Payment method is `cash` or `card`.
- "Chef accepts/declines an order" maps to `Confirm()` (accept) vs `Cancel()` (decline) — consider adding an explicit `declined` reason if product needs to distinguish.

---

## 5. Database plan

PostgreSQL, one table per migration, snake_case columns, soft-deletes via `is_active`, geo stored as `decimal` lat/lng with Haversine math (see `domain.CalculateDistance` and the `FindNearby` SQL in `repository/chef_repository.go`).

### Already migrated (`migrations/000001`–`000006`)

| Table | Purpose | Notes |
|---|---|---|
| `users` | accounts for customer / chef / admin | `role` column distinguishes them; has lat/lng for "deliver to me" |
| `chefs` | chef profile (1:1 with a user) | kitchen address + lat/lng + `delivery_radius`, `rating`, `is_accepting_orders` |
| `menus` | a chef's menu/collection | `menu_type`, availability schedule |
| `menu_items` | individual dishes/products | price, stock, dietary flags, `rating`, `total_orders` |
| `orders` | a customer order | pricing breakdown, `status`, `payment_method`, `payment_status`, delivery lat/lng |
| `order_items` | line items (snapshot) | **carries `chef_id`** → this is what enables one order spanning multiple chefs |

### Tables to add (to cover the full product brief)

These are **not built yet**. Add them as new numbered migrations following the existing style:

| Planned table | Why (feature) | Key columns |
|---|---|---|
| `password_reset_tokens` | "forgot password" | `user_id`, `token` (hashed), `expires_at`, `used_at` |
| `favorites` | customer favorites a chef | `user_id`, `chef_id`, unique(`user_id`,`chef_id`), `created_at` |
| `reviews` | rate chefs **and** products | `user_id`, `order_id`, nullable `chef_id`, nullable `menu_item_id`, `rating` (1–5), `comment` — feeds `chefs.rating` / `menu_items.rating` |
| `chat_conversations` | a thread between a customer and a chef | `user_id`, `chef_id`, `order_id?`, `last_message_at` |
| `chat_messages` | real-time messages | `conversation_id`, `sender_id`, `body`, `read_at`, `created_at` |
| `chef_earnings` (or derive) | chef "see earnings" | could be a view/aggregation over `order_items` joined to delivered/paid orders, or a ledger table with payout records |

Modeling decisions worth noting:
- **Chef online/offline:** the brief asks for online/offline. Today `chefs.is_accepting_orders` + `chefs.is_active` exist. Add a dedicated `is_online` boolean (presence) and keep `is_accepting_orders` for "open for business" — they answer different questions.
- **Best/popular products:** derive, don't store separately. "Popular" = order by `menu_items.total_orders` / `rating`; "best" = featured flag (`is_featured`) the chef sets.
- **Multi-chef orders:** keep order-level totals in `orders`, but because `order_items.chef_id` exists you can split an order per chef for chef-facing views and earnings. When you build chef order views, **filter `order_items` by `chef_id`**, not the whole order.
- **Earnings:** prefer computing from `order_items` where the parent order is `delivered`/`paid`. Only add a ledger table if you need payout/withdrawal tracking.

---

## 6. Current state (rebuilt from scratch)

The Go code was wiped and rebuilt as a clean, verified skeleton on branch
`rebuild/hexagonal-skeleton`. What exists today:

- **Runs end-to-end under Docker.** `docker compose up --build` starts Postgres + the API + Adminer; the API connects to the DB, runs all migrations via golang-migrate, and serves `GET /health` → `200 {"status":"ok","database":"ok"}`.
- **Implemented:**
  - `config`, `database` (connection + migration runner), `/health`, the router, the composition root.
  - **Authentication** — `User` entity + `UserRepository` port, Postgres adapter, `AuthService` (bcrypt + JWT via golang-jwt/v5), bearer-token middleware (`Auth.Require` / `RequireRole`), `POST /api/v2/auth/{register,login,logout}` + `GET /api/v2/auth/me`.
  - **Chef profiles** — `Chef` entity (+ Haversine `CanDeliverTo`) + `ChefRepository` port, Postgres adapter (incl. SQL Haversine `FindNearby`), `ChefService` (one profile per user, paging), `POST /api/v2/chefs` (auth) + `GET /api/v2/chefs`, `/chefs/nearby`, `/chefs/{id}`.
- **Schema vs. code:** `migrations/` defines `users`, `chefs`, `menus`, `menu_items`, `orders`, `order_items`. `users` and `chefs` are wired into Go; `menus`, `menu_items`, `orders`, `order_items` are not yet.
- **Tests already exist** — `go test ./...` runs green (`-race` clean). Domain, service and handler layers are covered. See §7a for the pattern to follow; **every new feature ships with tests.**
- **Not built:** forgot-password (needs a `password_reset_tokens` table + email), menus, orders, favorites, reviews, chat, earnings, online/offline toggle. Use these as the next features, each following the §2 recipe.

When asked to "implement X", build it inside-out with the §2 recipe (domain → port → service → repository → migration → handler → wire in `main.go` + `router`), and commit per feature on this branch.

---

## 7. Conventions

- **Layering is non-negotiable:** SQL only in `repository/`, business rules only in `domain/`+`service/`, HTTP only in `handler/`/`middleware/`. Wiring only in `cmd/api/main.go`.
- Repositories implement an interface declared in `domain/` and take `*sql.DB`; wrap errors with `fmt.Errorf("...: %w", err)`.
- Mutate entity state through domain methods (e.g. `order.MarkReady()`), not by assigning fields, so invariants/`updated_at` stay correct.
- Nullable DB columns → pointer fields (`*string`, `*float64`) in domain structs; `PasswordHash` is always cleared (`""`) before returning a user.
- API is versioned under `/api/v2`.
- Keep code `gofmt`-clean and `go vet`-clean — CI (`.github/workflows/test.yml`) enforces both, plus `go test -race`.

## 7a. Testing (this project already has tests — keep adding them)

Tests live next to the code as `*_test.go`. The hexagonal layering is what makes
this cheap: services depend on repository **interfaces**, so tests inject an
**in-memory fake repository** instead of touching Postgres. Established pattern,
already in the repo — copy it for every new feature:

- **Domain tests** (`internal/domain/*_test.go`): pure, no deps — defaults, validation, entity rule methods.
- **Service tests** (`internal/service/*_test.go`): black-box `package <pkg>_test`, a fake implementing the domain port, **table-driven** cases for success + each error path. See `auth_service_test.go` (`fakeUserRepo`).
- **Handler tests** (`internal/handler/*_test.go`): drive the **real router + middleware** with `httptest` over a fake repo; assert status codes and that secrets (e.g. password hash) never leak. See `auth_handler_test.go`.

Run: `go test ./...`, or `go test -race -cover ./...`. **A feature is not done until its tests are green.**

## 8. Common commands

```bash
make run            # run locally (go run ./cmd/api), needs Postgres + env
make dev            # full stack via docker-compose.dev.yml (hot reload via air)
make build          # build binary to bin/
make test           # go test ./...
make migrate-up     # apply migrations   (DB_URL overridable)
make migrate-down   # roll back last migration
make migrate-create name=create_favorites_table   # scaffold a new migration
make fmt            # go fmt ./...
```

App auto-runs migrations on boot when `AutoMigrate` is enabled (see `cmd/api/main.go`). Config comes from `.env*` files (see `.env.example`).
