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

> The code was rebuilt from scratch into a clean, Dockerized hexagonal service, and the **entire product brief above is now implemented** and released as `v3.0.0` (see §6 for the per-feature breakdown and §9 for versioning), including transactional email (`domain.Mailer`). When extending the app, keep following the §2 recipe.

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

Every layer is now populated (`domain/`, `service/`, `repository/`, `handler/`,
`middleware/`, `router/`); each package still keeps a `doc.go` stating its
hexagonal contract. `internal/middleware/` holds JWT auth plus the cross-cutting
logging, CORS and rate-limit middleware.

**Frontend:** `web/` is a separate **Vue 3 + Vite + JavaScript + Tailwind** SPA
(Pinia + Vue Router) — a single role-based app (customer + chef views). It is an
independent driving adapter: it only talks to the API over `/api/v2`, owns no
business rules, and has its own toolchain (`cd web && npm install`). Vite proxies
`/api` to the Go backend in dev. See `web/README.md`.

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
  - **Chef profiles** — `Chef` entity (+ Haversine `CanDeliverTo`) + `ChefRepository` port, Postgres adapter (incl. SQL Haversine `FindNearby`), `ChefService` (one profile per user, paging), `POST /api/v2/chefs` (chef role) + `GET /api/v2/chefs`, `/chefs/nearby`, `/chefs/{id}`, and `GET /api/v2/chefs/me` (chef role; 404 = no profile yet, drives UI onboarding).
  - **Menus & dishes** — `Menu` / `MenuItem` entities (+ stock/orderable rules) + `MenuRepository` / `MenuItemRepository` ports, Postgres adapters (soft-delete via `is_active`), `MenuService` (chef-owned CRUD with ownership checks; `MenuItem.chef_id` stamped from the menu). Chef-only mutations `POST/PUT/DELETE /api/v2/menus` and `/api/v2/menu-items`; public reads `GET /api/v2/menus/{id}`, `/menus/{id}/items`, `/chefs/{id}/menus`, `/chefs/{id}/menu-items`.
  - **Orders** — `Order` entity (status state machine per §4; `Cancel` only from pending/confirmed; `MarkPaid`/`Refund` payment machine) + `OrderItem` (price/name snapshot, carries `chef_id` for multi-chef carts) + `OrderRepository` port, Postgres adapter (order+items in one transaction; chef views filter `order_items` by `chef_id`). `OrderService`: place (validation, price calc, atomic stock decrement via `MenuItemRepository.DecrementStock`), customer get/list/cancel (ownership), chef list (scoped) + `AdvanceForChef` (confirm/preparing/ready/delivering/delivered/decline, gated on participation). Delivering a **cash** order settles it to `paid` (`Order.SettleCashOnDelivery`), so it counts toward earnings.
  - **Card payments (iyzico)** — hexagonal `PaymentGateway` port (`internal/payment`): the **iyzico** adapter (hosted Checkout Form, IYZWSv2 HMAC auth) when `IYZICO_API_KEY` is set, else a **dev mock** that simulates the dance via the SPA's `/mock-pay` page. Flow: `POST /api/v2/orders/{id}/pay` (owner) → hosted page → browser callback `POST /api/v2/payments/callback` (public, rate-limited) → **server-to-server verify** → `MarkPaid` + `payment_sessions` row (migration 000014). Cancelling a paid card order refunds through the gateway first (`OrderService` ← `domain.PaymentRefunder`); refund failure aborts the cancel. Config fails fast outside dev without credentials. Customer `POST/GET/cancel /api/v2/orders` (auth); chef `GET /api/v2/chef/orders` + `POST /api/v2/chef/orders/{id}/status` (chef role). Lifecycle/stock errors → 422.
  - **Favorites** — `Favorite` entity + `FavoriteRepository` port, Postgres adapter (idempotent `Add` via `ON CONFLICT DO NOTHING`; `ListChefs` reuses the shared chef columns/scanner). `FavoriteService` validates the chef exists then favorites. `POST/DELETE /api/v2/favorites/{chefId}` + `GET /api/v2/favorites` (auth).
  - **Reviews & ratings** — `Review` entity (+ `Validate`: rating 1–5, exactly one of chef/dish) + `ReviewRepository` port. `ReviewService.Create` enforces "review only what you actually received": order owned by the caller, target in the order, and the **target chef's sub-order delivered** — in a multi-chef order a chef who delivered is reviewable immediately (order-level status may still be pending) and a declined chef (and their dishes) never is. The Postgres adapter inserts the review and **recomputes `chefs.rating` / `menu_items.rating`** (`ROUND(AVG,2),COUNT`) in the same transaction; unique`(user_id,order_id,target)` → `ErrReviewExists`. `POST /api/v2/reviews` (auth), **rating history** `GET /api/v2/orders/{id}/reviews` (auth; user-scoped query — foreign orders yield an empty list), public `GET /api/v2/chefs/{id}/reviews`, `/api/v2/menu-items/{id}/reviews`. The SPA review panel loads the history and shows already-given ratings read-only; targets are limited to delivered slices.
  - **Online/offline** — `chefs.is_online` (migration 000009) + `Chef.SetOnline`; `ChefRepository.SetOnline` and an `onlineOnly` filter on `List`/`FindNearby`. `PATCH /api/v2/chefs/me/status` (chef); `?online=true` on `/chefs` + `/chefs/nearby`.
  - **Away / vacation mode (#104)** — gives the previously-dead `chefs.is_accepting_orders` column teeth (no new migration): `Chef.SetAcceptingOrders` + `ChefRepository.SetAcceptingOrders`; `ChefService.SetAcceptingOrders` toggles the caller's profile. **Distinct from online presence** (a soft `?online=true` filter) — when a chef is *not accepting orders* they are **hidden from browse/search** (hard `AND is_accepting_orders = true` in `List`/`FindNearby`/`SearchChefs`) and `PlaceOrder` rejects new orders for them (`ErrChefUnavailable` → 422, checked before stock decrement; separate from `ErrChefClosed` working-hours). `FindByID` still resolves an away chef so a direct/favorited link can show a badge. `PATCH /api/v2/chefs/me/availability` (chef, `{accepting_orders}`). SPA: a "go on vacation / reopen" toggle + 🌴 badge on the chef dashboard. TR + EN.
  - **Working hours (#70)** — `chef_hours` (migration 000020; minutes-since-midnight windows per weekday, `opens>closes` wraps past midnight, **no rows = always open**). `domain.ChefHours` + pure `IsOpenAt` (opens inclusive, closes exclusive; overnight + split-shift windows); evaluated in the platform TZ (`TIMEZONE`, default Europe/Istanbul, UTC fallback with a warning). `ChefHoursRepository` (full-replace tx + batched `ListByChefs`). **`PlaceOrder` rejects orders when any participating chef is closed** (`ErrChefClosed` → 422). `PUT /api/v2/chefs/me/hours` (chef, HH:MM wire format, full replace, empty clears) + public `GET /api/v2/chefs/{id}/hours`; chef reads (`Get`/`List`/`Nearby`) carry a derived `is_open_now`. SPA: weekly editor on the profile, "closed now" badges on browse/detail.
  - **Chef earnings** — derived (no ledger): `EarningsRepository` sums the chef's **delivered slices on paid orders** (sub-order gate), optional `since` window. `GET /api/v2/chefs/me/earnings?days=N` (chef) returns gross food subtotal, delivery fees (kept in full), **tips** (kept in full), commission and **net** (gross + delivery + tips − commission).
  - **Tips (#105)** — an optional customer gratuity at checkout (`tip` on `POST /api/v2/orders`; negative → 400). Added to the order total (so iyzico `paidPrice` includes it, no gateway change) and paid to the chef **uncommissioned**. Snapshotted on `orders.tip` and, for multi-chef carts, **split across sub-orders proportionally to food subtotal** (`domain.DistributeTip` — cents-rounded, remainder on the largest slice, sum exact) onto `sub_orders.tip` (migration 000024). Feeds chef earnings; declining a card-paid slice refunds food **+ delivery fee + tip share**. SPA: percentage-preset + custom tip selector in the cart (grand total), tip line on orders, tips line in the earnings breakdown. TR + EN.
  - **Promo codes (#94)** — `promo_codes` (migration 000022; percent/fixed, `min_order`, validity window, `usage_limit`/`used_count`) + `orders.promo_code` snapshot. `PromoCode` domain (`Validate`, `Redeemable`, `DiscountFor` capped at subtotal), `PromoRepository` (case-insensitive `FindByCode`; **atomic `Redeem`** — guarded `UPDATE … WHERE used_count < usage_limit` makes the cap race-free, proven by a 20-goroutine integration test). `PlaceOrder` accepts `promo_code`, validates + redeems + snapshots `orders.discount`/`promo_code`, recomputes the total; **platform-funded** — the discount is off the order total, chef earnings (from undiscounted sub-order subtotals) are untouched. Admin CRUD: `GET/POST /api/v2/admin/promos` + `PATCH /…/{id}/active` (via `AdminService`). Errors map to 404 (unknown) / 422 (expired/used-up/below-min) / 409 (duplicate) / 400 (bad definition). SPA: promo field in the cart, discount line on orders, promo tab in the admin panel. TR + EN.
  - **Money model (#65)** — `domain.FeePolicy` from config (`DELIVERY_BASE_FEE`, `DELIVERY_FEE_PER_KM`, `COMMISSION_PERCENT`; zeros = free, fail-fast on garbage/out-of-range). Customer pays a **distance-based delivery fee per chef slice** (base + per-km Haversine kitchen→address; base only without coordinates); the platform's **commission (% of the food subtotal) is deducted from the chef**, never charged to the customer. Both amounts are **snapshotted onto `sub_orders`** (`delivery_fee`, `commission`, migration 000021) at placement — rate changes never rewrite history. Declining a card-paid slice refunds food **+ its delivery fee**; `RoundMoney` keeps totals dust-free; iyzico needs no change (`paidPrice` > basket `price` is the documented fee pattern). UI: cart hint, delivery-fee line on orders, net-earnings breakdown on the dashboard.
  - **Order ETAs (#92)** — when a chef accepts (`confirm`), `Order.SetEstimatedDelivery` stamps `estimated_delivery_time = now + ETA_MINUTES` **once** (idempotent — a later chef confirming in a multi-chef order doesn't move it); `OrderService.SetETAWindow` from config (`ETA_MINUTES`, default 45, 0 disables). Persisted via `updateOrderRow`; the SPA order view shows the estimate + a "in N min" hint while the order is in progress.
  - **Dietary tags & filters (#91)** — `menu_items` dietary flags (vegetarian/vegan/gluten-free/halal/spicy + calories) already round-trip through the chef menu editor; dish search adds bound-bool filters (`?vegetarian=&vegan=&gluten_free=&halal=` on `GET /api/v2/search?type=food`, each `(NOT $n OR is_flag)` — parameterized, index-friendly). SPA: dietary checkboxes in the menu editor, `DietaryBadges` on dish cards (detail/search/menus), filter chips on the search dish tab. TR + EN.
  - **Search + filters/sorting (#68)** — `SearchRepository` ILIKE over chefs/dishes/users (index-backed by pg_trgm GIN, migration 000010). `GET /api/v2/search?q=&type=&limit=&offset=` (auth) now also takes `min_rating`, `min_price`/`max_price`/`cuisine` (dishes) and `sort` (`rating|popular|price_asc|price_desc`, price for dishes only); `GET /api/v2/chefs` takes `min_rating` + `sort` (`domain.ChefListFilters`). Sort values are a **whitelist mapped to fixed ORDER BY expressions** (`chefOrder`/`dishOrder` in the repo — never interpolated) and every ordering ends with an `id` tiebreaker for stable pagination; the service rejects unknown sorts/out-of-range values with 400. SPA: filter bar + sort dropdowns on browse and search. type user stays admin-only.
  - **Forgot/reset password** — `password_reset_tokens` (migration 000011, stores only sha256 of the token); `PasswordResetRepository`; `AuthService.RequestPasswordReset` (single-use, expiring, silent for unknown emails) + `ResetPassword`. `POST /api/v2/auth/forgot-password` (always 202) + `/reset-password`. The reset link is delivered via the `domain.Mailer` port (`internal/mailer`: real SMTP, or a dev logging mailer when `SMTP_HOST` is empty); the token is **never** returned in an API response.
  - **Email verification** — `email_verification_tokens` (migration 000023, sha256-only, single-use, 24h expiry, mirrors the reset flow); `EmailVerificationRepository` + `UserRepository.MarkVerified`. Registration issues a token and emails a `…/verify-email?token=` link via `domain.Mailer` (non-fatal — a send failure only logs, the account still exists); redeeming it flips `users.is_verified`. `POST /api/v2/auth/verify-email` (public, rate-limited) + `POST /api/v2/auth/resend-verification` (auth, rate-limited, 409 once already verified). Verification is **surfaced, not enforced**: `/auth/me` carries `is_verified` and the SPA shows a resend banner (`web/src/components/VerifyBanner.vue`) — ordering is not gated on it. The raw token is never returned in an API response.
  - **Photo upload (#63)** — hexagonal `domain.FileStore` port; `internal/storage.Local` disk adapter (random hex names, extension whitelist — client filenames never touch the filesystem, so traversal is impossible by construction; `storage.ValidName` gates serving). `service.UploadService` **decodes and re-encodes** every upload (proves it's a real JPEG/PNG regardless of claimed type, strips EXIF/GPS and trailing payloads) and enforces ownership (chef's own dish / own kitchen). `POST /api/v2/menu-items/{id}/image` + `POST /api/v2/chefs/me/image` (chef role, multipart `image`, 5 MB cap → 413) → `{image_url}`; public `GET /uploads/{file}`. `chefs.image_url` (migration 000019; dishes had it since 000004). `UPLOAD_DIR` config; `uploads_data` volume in dev+prod compose (Dockerfile pre-creates `/app/uploads` owned by the app user — first-mount ownership copy makes the volume writable for non-root); Caddy + Vite proxy `/uploads`. SPA: per-dish photo upload in My menus, kitchen photo in profile; images on browse/detail/search cards. **Multi-photo galleries (#93):** the same pipeline appends to `menu_items.images` (JSON array, cap `MaxGalleryImages`=5) via `POST/DELETE /api/v2/menu-items/{id}/images` (`UploadService.Add/RemoveDishGalleryImage`, owner-only, `ErrGalleryFull`→422); `MenuItemRepository.SetImages`. My menus manages the gallery (thumbnails + add/remove); dish detail shows the gallery strip.
  - **Data rights / GDPR (#107)** — no migration. **Export**: `GET /api/v2/users/me/export` (auth) returns a caller-scoped JSON dump (`domain.AccountExport`: account with the hash cleared, chef profile, addresses, orders they placed, reviews they wrote, chat threads they took part in) as a `Content-Disposition: attachment` download; `AccountService.Export` orchestrates existing read ports — never another user's data. **Delete**: `DELETE /api/v2/users/me` (auth, password-confirmed → 401 on mismatch) **anonymises** rather than hard-deletes (hard delete would cascade away the customer's orders → chefs' earnings/history, and is FK-blocked for chefs in any order). `AccountRepository.Anonymise` (one tx): scrub `users` PII (`email`/`username` → `deleted-<id>…` so UNIQUE holds and the address frees up, phone/location NULL, `password_hash=''`, `is_active=false`), scrub the chef storefront (business name → "Closed kitchen", location/certs cleared, offline + not accepting), delete addresses, tombstone chat message bodies to `[deleted]`, drop reset/verification tokens — **orders/order_items/sub_orders/reviews retained** for counterparties. The handler revokes the presented token on success (like logout); login is then blocked (`is_active=false`). `ReviewRepository.ListByUser` added for the export. SPA: "Data & privacy" card on `/profile` (export download + password-confirmed delete → logout). TR + EN.
  - **Address book (#66)** — `addresses` (migration 000018; partial unique index enforces one default per user, FK cascades on user delete). `Address` entity + `AddressRepository` port, Postgres adapter (default swap in a tx), `AddressService` (owner-only CRUD, first address auto-defaults), auth-only `GET/POST /api/v2/addresses` + `PUT/DELETE /api/v2/addresses/{id}`. `PlaceOrder` accepts `address_id` (mutually exclusive with `delivery_address`, owner-checked) and **snapshots** the text/city/coords onto the order — editing/deleting book entries never rewrites history. SPA: book management on `/profile`, saved-address selector (default preselected, "other" = free text) in the cart.
  - **Profile management (#64)** — logged-in **password change** (`AuthService.ChangePassword`: current password verified with bcrypt, `PUT /api/v2/auth/password`), **user profile edit** (`PUT /api/v2/users/me`: contact + default location only — email/username/role are immutable by design), **chef kitchen edit** (`PUT /api/v2/chefs/me`, chef role, resolved by the caller's user id). `users.phone_number` widened to varchar(20) (migration 000016 — formatted TR numbers overflowed 15). SPA `/profile` page (account + password + kitchen forms; navbar avatar chip links to it).
  - **Email notifications (#58, #71)** — `service.OrderNotifier` (reuses the `domain.Mailer` port): "new order" to each participating chef on placement (their slice only) and "status changed" to the customer on meaningful sub-order transitions (confirmed/delivering/delivered/declined — not preparing/ready). **Fire-and-forget**: goroutines with `context.WithoutCancel`, failures logged (`order email failed`), never surfaced to the buyer; nil notifier disables. No new config — SMTP in prod, dev logging mailer otherwise. **Per-user opt-out** (#71): `users.email_notifications` (migration 000017, default on) checked by the notifier for both chef and customer sends; toggled via `PUT /users/me` (`email_notifications`, omit = keep) and the SPA profile page. Password-reset email is *not* governed by the flag.
  - **Real-time chat** — `chat_conversations` / `chat_messages` (migrations 000012/000013), `Conversation` (+ `IsParticipant`) / `Message`, `ChatRepository`, `ChatService` (participant-only: customer by `user_id`, chef by chef profile). REST history + a WebSocket `/ws` endpoint (`gorilla/websocket`) driven by a concurrency-safe `Hub` in the handler layer; REST posts also broadcast live. `…/api/v2/chat/conversations[/{id}/messages|/ws]` (auth). Because browsers can't set headers on WS handshakes, the auth middleware accepts `?access_token=` **for upgrade requests only** (the request logger never records query strings). **Read receipts (#95):** `chat_messages.read_at`; `POST /…/conversations/{id}/read` (participant-only) marks the other party's messages read (`ChatRepository.MarkRead` — `sender_id <> reader`); the conversation list carries a derived `unread_count` per thread (computed in SQL from `c.user_id`, no extra param). SPA marks read on opening a thread + shows unread badges in the thread list. **Live "seen" (#106):** every outbound WS frame is a tagged envelope — `{type:"message",message}` for a new message, `{type:"read",read:{conversation_id,reader_id,read_at}}` broadcast when a participant marks the thread read — so the sender's connected client flips their delivered messages to "seen" without a refetch (inbound client→server frames are still the bare `{body}`). SPA unwraps by `type` and shows a single "seen" marker under the last read outgoing bubble. **WS upgrade fix:** the metrics middleware's status wrapper now forwards `Hijack` (it didn't since #73, which had been 500-ing every `/ws` handshake once metrics were in the chain).
  - **Admin panel (#69)** — `admin` role gets a moderation surface: `AdminRepository` (all admin SQL — cross-entity listings incl. inactive rows, `Stats` aggregation) + `AdminService`. `GET /api/v2/admin/{stats,users,chefs,orders}` and `PATCH /api/v2/admin/{users,chefs}/{id}/active`, all behind `RequireRole(admin)` (the role guard *is* the boundary — no per-user ownership). Deactivating a **user** blocks login (`AuthService.Login` already checks `is_active`); deactivating a **chef** hides them from browse/search (both filter `is_active`) and blocks new orders (the active-only `FindByID` in `PlaceOrder` → 404 on a stale cart). Stats: GMV (delivered & paid), orders/day, top chefs by delivered revenue — pure SQL, no new tables. Admin can't self-deactivate (422). Password hashes cleared in listings. First admin seeded out of band (`make seed-admin EMAIL=…` → `UPDATE users SET role='admin'`), never via the API. SPA `/admin` behind a `role:'admin'` router guard (overview/users/chefs/orders tabs).
  - **Authorization** — privileged `admin` role can't be self-assigned at registration; chef-only endpoints are guarded by `RequireRole(chef)` (see `router.handleRole`); plain-auth endpoints use `router.handleAuth`; services additionally enforce per-resource ownership (`domain.ErrForbidden` → 403).
  - **Observability (#73)** — Prometheus `/metrics` (`internal/metrics`, private registry): HTTP `http_requests_total{method,status}` + `http_request_duration_seconds` (labelled by method/status only — **never raw paths**, cardinality guard), Go runtime + process + `database/sql` pool collectors, and `orders_placed_total`/`orders_delivered_total`/`payments_total{outcome}` business counters (incremented nil-safe in the order/payment handlers). The endpoint is mounted on a **top mux outside the logging + metrics middleware** (scrapes don't self-count or spam logs) and only on the API's internal port — Caddy 404s `/metrics` publicly, Prometheus scrapes `api:8080` on the Docker network. `--profile monitoring` in prod compose adds Prometheus + Grafana (`deploy/monitoring/`) with a starter dashboard and alert rules; both localhost-bound.
  - **Cross-cutting hardening** — JWTs carry a `jti`; logout (authenticated) revokes it via a `service.TokenRevoker` checked in the auth middleware. Per-IP rate limiting (`middleware.Limiter`, 429) on the auth endpoints + payment callback + the password-bearing authenticated endpoints (change-password, delete-account — `router.limitedAuth` throttles *and* requires a session, so a stolen token can't brute-force the account password; #115). Both are **in-memory by default and Redis-backed when `REDIS_URL` is set** (`internal/redisstore`; shared across instances, fail-open on Redis errors). Global middleware: structured request logging (`log/slog` + `middleware.RequestLogger`, X-Request-ID) and `middleware.CORS` (honours `ALLOWED_ORIGINS`). `cmd/api/main.go` runs an `http.Server` with timeouts + SIGINT/SIGTERM graceful shutdown. Config requires `JWT_SECRET` and rejects the placeholder outside development. List endpoints return a `{data,limit,offset,total}` envelope (`respondPage`).
- **Schema vs. code:** `migrations/` 000001–000024 define `users`, `chefs` (+`is_online`), `menus`, `menu_items`, `orders` (+`tip`), `order_items`, `favorites`, `reviews`, `password_reset_tokens`, `chat_conversations`, `chat_messages`, `payment_sessions`, `sub_orders` (+`tip`), `promo_codes`, `email_verification_tokens`, plus pg_trgm search indexes — all wired into Go.
- **Tests** — `go test ./...` runs green (`-race` clean) for domain/service/handler over fakes; `make test-integration` runs the repository adapters + migrations against a real Dockerized Postgres (`-tags=integration`). See §7a; **every new feature ships with tests.**
- **Per-chef sub-orders (#34):** `sub_orders` (migration 000015; one row per order+chef, backfilled from existing orders) gives each chef's slice its own §4 lifecycle (`domain.SubOrder`); the order-level `status` is **derived** (`DeriveOrderStatus`: all cancelled → cancelled, all active delivered → delivered, else least-advanced active). `AdvanceForChef` moves only the caller's sub-order and persists it with the re-derived parent atomically (`OrderRepository.UpdateSubOrder`, `FOR UPDATE` on the order row). Declining a slice of a **card-paid** order partial-refunds that chef's subtotal (`PaymentRefunder.RefundSubOrderPayment` → `PaymentGateway.RefundPartial`; iyzico `/v2/payment/refund`) and a failed refund aborts the decline. Customer cancel needs **every** sub-order still pending/confirmed and cancels them all. Earnings count a chef's slice once *their* sub-order is delivered (order paid); cash still settles order-level when everything is delivered. Chef UI badges/actions run off the caller's sub-order; customer UI shows per-chef progress chips on multi-chef orders.
- **Features + hardening: complete.** All feature issues (#1–#9), the tech-debt/security issues (#10–#12, #14–#18), and email delivery (#20, `domain.Mailer` + `internal/mailer`) are implemented. The remaining follow-up is operational: for multi-instance deploys, move the token denylist + rate limiter to a shared store like Redis.

When asked to "implement X", build it inside-out with the §2 recipe (domain → port → service → repository → migration → handler → wire in `main.go` + `router`), and commit per feature on this branch.

---

## 7. Conventions

- **Layering is non-negotiable:** SQL only in `repository/`, business rules only in `domain/`+`service/`, HTTP only in `handler/`/`middleware/`. Wiring only in `cmd/api/main.go`.
- Repositories implement an interface declared in `domain/` and take `*sql.DB`; wrap errors with `fmt.Errorf("...: %w", err)`.
- Mutate entity state through domain methods (e.g. `order.MarkReady()`), not by assigning fields, so invariants/`updated_at` stay correct.
- Nullable DB columns → pointer fields (`*string`, `*float64`) in domain structs; `PasswordHash` is always cleared (`""`) before returning a user.
- API is versioned under `/api/v2`.
- Keep code `gofmt`-clean and `go vet`-clean — CI (`.github/workflows/test.yml`) enforces both, plus `go test -race`.
- **Docs ship with the code (same PR):** `README.md`/`web/README.md` for user-visible features and commands, this file (§6 current state per feature, §9 tag history per release, §10 when the security posture moves), `SECURITY.md` when auth/payment behaviour changes, `DEPLOY.md` for anything operational, `CONTRIBUTING.md` when the workflow changes. A PR that changes behaviour but leaves the docs stale is incomplete.

## 7a. Testing (this project already has tests — keep adding them)

Tests live next to the code as `*_test.go`. The hexagonal layering is what makes
this cheap: services depend on repository **interfaces**, so tests inject an
**in-memory fake repository** instead of touching Postgres. Established pattern,
already in the repo — copy it for every new feature:

- **Domain tests** (`internal/domain/*_test.go`): pure, no deps — defaults, validation, entity rule methods.
- **Service tests** (`internal/service/*_test.go`): black-box `package <pkg>_test`, a fake implementing the domain port, **table-driven** cases for success + each error path. See `auth_service_test.go` (`fakeUserRepo`).
- **Handler tests** (`internal/handler/*_test.go`): drive the **real router + middleware** with `httptest` over a fake repo; assert status codes and that secrets (e.g. password hash) never leak. See `auth_handler_test.go`.
- **E2E smoke** (`web/e2e/*.spec.js`, Playwright): one golden-path browser test (chef onboarding → menu → order → deliver → cash settles) against the real SPA + API + Postgres; own CI job (`e2e`). Run locally with the stack up: `cd web && npm run test:e2e`.
- **Repository integration tests** (`internal/repository/*_integration_test.go`): the only tests that hit a **real Postgres** — they exercise the actual SQL (Haversine `FindNearby`, decimal scanning, the order transaction + multi-chef scoping, `ON CONFLICT` idempotency, atomic `DecrementStock`) and the migration files. Gated behind the **`//go:build integration`** tag, so the default `go test ./...` skips them and needs no database. `TestMain` (`main_test.go`) connects via `TEST_DATABASE_URL`, runs migrations, and shares `testDB`; `helpers_test.go` provides `resetDB` (TRUNCATE … RESTART IDENTITY CASCADE) and seed helpers. The DB is a throwaway `postgres:16-alpine` from `docker-compose.test.yml` (host port 5433); CI runs them in a separate job with a Postgres service container.

Run: `go test ./...` (unit, no DB), `go test -race -cover ./...`, or `make test-integration` (spins up Dockerized Postgres, runs the tagged suite, tears down). **A feature is not done until its tests are green** — repository work should ship an integration test.

## 8. Common commands

```bash
make run            # run locally (go run ./cmd/api), needs Postgres + env
make dev            # full stack via docker-compose.dev.yml (hot reload via air)
make build          # build binary to bin/
make test           # go test ./...  (unit + handler; no database needed)
make test-integration  # Dockerized Postgres + repository integration tests (-tags=integration)
make migrate-up     # apply migrations   (DB_URL overridable)
make migrate-down   # roll back last migration
make migrate-create name=create_favorites_table   # scaffold a new migration
make fmt            # go fmt ./...
```

App auto-runs migrations on boot when `AutoMigrate` is enabled (see `cmd/api/main.go`). Config comes from `.env*` files (see `.env.example`).

**Production** (`docker-compose.prod.yml` + `make prod`, runbook in `DEPLOY.md`): Caddy (`Dockerfile.web` + `deploy/Caddyfile`) serves the built SPA and proxies `/api`/`/health`/`/version`/`/uploads` to the Go API on **one origin** (no CORS; auto-HTTPS via `SITE_ADDRESS=<domain>`); DB and API have no host ports. `.env.prod` (from `.env.prod.example`, gitignored) is enforced by compose `${VAR:?}` **and** the API's `ENV=production` fail-fasts. **Backups (#74):** the `db-backup` sidecar (postgres:16-alpine + `deploy/backup/`) takes a nightly `pg_dump -Fc` + uploads tarball into `./backups/` on the host (also once at startup), pruned after `BACKUP_RETENTION_DAYS` (default 14); restore runbook + drill procedure in `DEPLOY.md`, executed drill recorded on issue #74. Off-site copying of `./backups/` is an operator step.

---

## 9. Versioning & releases

Releases are tracked with **annotated git tags following SemVer** (`vMAJOR.MINOR.PATCH`). There are **no long-lived version branches** — work happens on short-lived feature branches that merge into `main`, and each release is a tag on `main`. (The old `archive/v1` / `v2` branches were deleted; their history is preserved in `main` and the tags below.)

Bump rules:
- **MAJOR** — incompatible API/behaviour change (or a ground-up rewrite).
- **MINOR** — backwards-compatible feature.
- **PATCH** — backwards-compatible fix.

Tag history:

| Tag | What it marks |
|---|---|
| `v0.1.0` | early "move to DDD" of the original project |
| `v1.0.0` | original project, documented |
| `v2.0.0` | original gin-based v2 (pre-rebuild) |
| `v3.0.0` | the hexagonal rebuild — full product brief + hardening (PR #19, closes #1–#18) |
| `v3.1.0` | email delivery — `domain.Mailer` (SMTP + dev logging); password reset emails the link (PR #21, closes #20) |
| `v3.2.0` | web UI feature-complete — the Vue SPA covers the entire product brief incl. real-time chat; config tests at 100%; WS query-token auth (closes #24, #26–#31, #36) |
| `v3.3.0` | money flow complete — cash settles on delivery; iyzico card checkout via the `PaymentGateway` port + refund-on-cancel; `/version`; Vitest in CI (closes #33, #42, #43) |
| `v4.0.0` | **production release** — one-origin deploy behind Caddy with auto-HTTPS (`DEPLOY.md`); Redis-backed denylist + rate limiter for multi-instance (closes #32, #44) |
| `v4.1.0` | per-chef sub-order status — `sub_orders` + derived order status, partial refunds on decline, per-slice earnings; TR/EN web UI incl. i18n (PR #60/#61, closes #56, #34) |
| `v4.2.0` | order email notifications — new order → chef, sub-order status changes → customer; fire-and-forget over `domain.Mailer` (PR #62, closes #58) |
| `v4.3.0` | account self-service — profile management (#64), email-notification opt-out (#71), customer address book (#66); SECURITY.md + CONTRIBUTING.md |
| `v4.4.0` | review history + per-slice reviewability (PR #81); search filters & whitelisted sorting (#68, PR #82) |
| `v4.5.0` | photo upload — dish + kitchen photos via the `FileStore` port, EXIF-stripping re-encode, traversal-proof serving (#63, PR #83) |
| `v4.6.0` | chef working hours — weekly schedules (overnight + split shifts), order-time enforcement in Europe/Istanbul, `is_open_now` (#70, PR #84) |
| `v4.7.0` | money model — distance-based delivery fees + chef commission with per-slice snapshots (#65, PR #86); nightly DB backups + restore runbook (#74, PR #85) |
| `v4.8.0` | CD pipeline: build/push to GHCR + deploy on tag (#72); observability: Prometheus `/metrics`, Grafana dashboard, alerts (#73) |
| `v4.9.0` | staging environment: production-shaped throwaway stack, `ENV=staging` mock relaxation, deploy-on-main CD, E2E via `E2E_BASE_URL` (#75, PR #89) |
| `v4.10.0` | admin panel: platform stats, user/chef activation, order overview behind `RequireRole(admin)` (#69, PR #90) |
| `v4.11.0` | dietary tags & filters (#91); reorder from history (#96) |
| `v4.12.0` | order ETAs (#92); multi-photo dish galleries (#93) |
| `v4.13.0` | chat read receipts + unread badges (#95, PR #101) |
| `v4.14.0` | discount/promo codes: percent/fixed, platform-funded, atomic usage cap, admin management (#94, PR #102) |
| `v4.15.0` | email verification on registration — `email_verification_tokens`, verify/resend endpoints, `is_verified` surfaced (not enforced) with an SPA resend banner (#103, PR #108) |
| `v4.16.0` | chef away / vacation mode — one-tap availability toggle giving `is_accepting_orders` teeth (hidden from browse/search + orders blocked), no migration (#104, PR #109) |
| `v4.17.0` | live chat "seen" receipts over WebSocket — tagged message/read frame envelope; fixes the metrics-middleware Hijack bug that had 500'd `/ws` since #73 (#106, PR #110) |
| `v4.18.0` | tips for chefs at checkout — order-level gratuity split per sub-order by subtotal, uncommissioned, into earnings + refunds (#105, PR #111) |
| `v4.19.0` | account deletion + data export (GDPR) — caller-scoped JSON export; password-confirmed delete that anonymises (retaining counterparty orders/reviews) and revokes the token (#107, PR #112) |
| `v4.19.1` | security fix: anonymised delete identity now carries a random suffix so a pre-registered `deleted-<id>` sentinel can't block a user's deletion (#113, PR #114) |

Cutting a release (annotated tag on a clean, green `main`):

```bash
git checkout main && git pull
git tag -a vX.Y.Z -m "vX.Y.Z — <summary>"
git push origin vX.Y.Z
```

`git describe --tags` feeds the binary version via ldflags (`GET /version`). **Pushing the tag also triggers CD** (`.github/workflows/release.yml`, #72): it builds + pushes the API/web images to GHCR stamped with the tag and, when `DEPLOY_ENABLED=true` + the `DEPLOY_*` secrets are set, rolls the production host to that tag over SSH (pull + `up -d --no-build` + `/health` smoke check). `docker-compose.prod.yml` selects the image by `${TAG}`; `make prod` still builds locally. Rollback = push/deploy an older tag. **Staging (#75):** pushes to `main` build `:main` images (`.github/workflows/staging.yml`) and deploy to a staging host when `STAGING_ENABLED=true`; `docker-compose.staging.yml` mirrors prod's topology but `ENV=staging` permits the mock gateway + dev mailer (config's `allowsMocks`) while keeping the strict JWT secret, on isolated volumes/network + ports 8090/8453 (`make staging`). The E2E smoke runs against any environment via `E2E_BASE_URL`. See `DEPLOY.md`.

---

## 10. Security policy (OWASP Top 10 analysis)

Security posture is tracked against the **OWASP Top 10 (2021)**. The table maps each category to this codebase's mitigations; the policies below it are **binding for all new code**.

| OWASP | Where this project stands |
|---|---|
| **A01 Broken Access Control** | Role guards at the router (`handleAuth` / `handleRole(chef)`); **per-resource ownership enforced in services** (`domain.ErrForbidden` → 403): chefs own their menus/dishes, customers their orders/reviews, chat is participant-only. `admin` cannot be self-assigned at registration. |
| **A02 Cryptographic Failures** | Passwords: bcrypt. Reset tokens: only the **sha256** stored; raw token never in an API response. JWTs: HS256 with a required strong secret; `PasswordHash` cleared before any response. |
| **A03 Injection** | `database/sql` with **parameterized placeholders (`$1…`) only** — SQL lives exclusively in `internal/repository/`. Search uses bound `ILIKE` args, never concatenated input. |
| **A04 Insecure Design** | State machines for order/payment transitions (illegal moves rejected); single-use, expiring reset tokens; forgot-password is silent for unknown emails (no account enumeration); price/name **snapshots** on order items. |
| **A05 Security Misconfiguration** | Config **fails fast**: missing/placeholder `JWT_SECRET` outside dev, missing SMTP outside dev. No committed secrets (compose reads env). CORS allowlist via `ALLOWED_ORIGINS`; HTTP server timeouts + graceful shutdown. |
| **A06 Vulnerable Components** | CI runs `npm audit` (production deps, high+) for the web app; Go deps are version-pinned via modules. Keep dependencies current. |
| **A07 Identification & Auth Failures** | Per-IP **rate limiting** on register/login/forgot/reset/verify/resend **and the password-bearing authenticated endpoints** (change-password, delete-account) (429) — a stolen session can't become a brute-force oracle; logout revokes the token's `jti` via the denylist; login errors are generic (`invalid credentials`). |
| **A08 Software & Data Integrity** | Protected flow: feature branch → PR → **green CI** (vet, gofmt, `-race` tests, Postgres integration, web build) → merge; versioned SQL migrations. |
| **A09 Logging & Monitoring Failures** | Structured `slog` request logs (request id, status, latency); unexpected 500s logged server-side. Prometheus `/metrics` (#73, `internal/metrics`) — HTTP rate/latency/status, Go+process+DB-pool, order/payment counters — served on the internal port only (Caddy 404s the public path); `monitoring` compose profile adds Prometheus + Grafana + alert rules (5xx, p95, DB, API-down). **Never log credentials, JWTs or reset tokens** — the dev logging mailer prints reset links *in development only* by design. |
| **A10 SSRF** | The API makes no user-controlled outbound requests (SMTP target is operator config). Any future webhook/URL-fetch feature must allowlist destinations. |

**Policies for new code (non-negotiable):**
1. Every new endpoint declares its access level in the router (`public` is an explicit choice, not a default) and enforces **ownership in the service layer**, never only in the handler.
2. SQL only in `repository/` and only with placeholders — string-built SQL with user input never passes review.
3. Secrets never in code, compose files, fixtures or logs; new required config must fail fast outside development.
4. New auth-adjacent endpoints (anything unauthenticated that writes or reveals account state) get rate limiting and enumeration-safe responses.
5. Every feature ships with tests for its authorization paths (401/403 cases), per §7a.
6. Run `/security-review` on branches touching auth, payments, file handling or new dependencies before opening the PR.
