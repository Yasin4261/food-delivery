# Home Chef — Web UI

A single **role-based** Vue 3 SPA for the food-delivery API: the same app shows
customer or chef views depending on the logged-in user's role.

Stack: **Vue 3 + Vite + JavaScript + Tailwind CSS**, Pinia (state), Vue
Router and vue-i18n.

## Develop

The Go API must be running (e.g. `docker compose up` at the repo root, serving
`http://localhost:8080`). Vite proxies `/api/*` and `/health` to it, so the SPA
and API share an origin in dev — no CORS needed.

```bash
cd web
npm install
npm run dev        # http://localhost:5173
```

Point at a different backend with `VITE_API_TARGET=http://host:port npm run dev`.

## Test & build

```bash
npm test           # Vitest (jsdom): stores, router guards, api client, components
npm run test:watch # watch mode
npm run test:e2e   # Playwright golden-path smoke (needs the API running: docker compose up)
npm run build      # outputs dist/
npm run preview    # serve the production build
```

Unit tests live next to the code as `*.test.js` and run in CI (web job) before
the build. The **E2E smoke** (`e2e/shop-flow.spec.js`) drives a real browser
through register → chef onboarding → menu/dish → order → deliver → cash
settlement, against the real API + Postgres — it has its own CI job.

## Layout

```
src/
  api/client.js        # fetch wrapper: bearer token, list-envelope unwrap, 401 handling
  stores/auth.js       # token + user (Pinia), login/register/logout
  stores/cart.js       # multi-chef cart, persisted to localStorage
  router/index.js      # routes + auth/role guards
  i18n/                # vue-i18n setup + en/tr message catalogues
  lib/status.js        # order-status colours + chef transition actions
  components/NavBar.vue
  views/               # Login, Register, Chefs, ChefDetail, Cart, Orders, ChefDashboard
```

## What's covered

- Auth: register / login / logout (JWT in `localStorage`; 401 ⇒ auto logout);
  login shows first for anonymous visitors.
- Customer: browse chefs, find nearby, view a chef's menu, multi-chef cart,
  place an order (cash/card), see order history, cancel pending/confirmed.
- Chef: **onboarding** (create the kitchen profile when none exists), dashboard
  with earnings + online/offline toggle, incoming orders with status actions
  (accept → preparing → ready → delivering → delivered, or decline), and
  **My menus** — create/delete menus and add/remove dishes (price, stock or
  unlimited).

- Favorites: heart toggle on browse/chef detail and a Favorites page.
- Reviews: rate delivered orders (chef + dishes, stars + comment) from My
  orders; chef detail shows the reviews list and per-dish ratings.
- Search: `/search` with chef ↔ dish tabs over `GET /api/v2/search`; add to
  cart straight from dish results.
- Password reset: "Forgot password?" on login → emailed link opens
  `/reset-password?token=…` (in dev the link is printed by the API's logging
  mailer — `docker logs food_delivery_api`).
- Chat: `/chat` — conversation list + live thread over WebSocket ("Chat with
  chef" on the chef page; chefs answer from the same screen). The WS handshake
  authenticates via `?access_token=` (browsers can't set headers there);
  falls back to REST posting when the socket is down.
- Card payments: "💳 Pay now" on pending card orders → hosted checkout
  (iyzico in production; in dev the mock gateway sends you to `/mock-pay`, a
  simulated sandbox page) → redirected back to `/orders` with a result banner.
  Payment badges on every order; paid card orders refund on cancel.

- Profile: `/profile` (navbar avatar chip) — edit contact/default location,
  change password (current password required), toggle order-email
  notifications, and for chefs the kitchen profile (name, specialty, address,
  radius, coordinates).
- Localisation: full **Turkish + English** UI (vue-i18n). The navbar switcher
  toggles TR/EN; the choice persists in `localStorage` and first visits
  language-detect from the browser (Turkish browsers get Turkish). Catalogues
  live in `src/i18n/{en,tr}.js` — note vue-i18n treats `@` and `|` as syntax,
  so literal `@` in a message must be written `{'@'}`.

The full product brief is covered in the UI. 🎉
