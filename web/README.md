# Home Chef — Web UI

A single **role-based** Vue 3 SPA for the food-delivery API: the same app shows
customer or chef views depending on the logged-in user's role.

Stack: **Vue 3 + Vite + JavaScript + Tailwind CSS**, Pinia (state) and Vue
Router.

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

## Build

```bash
npm run build      # outputs dist/
npm run preview    # serve the production build
```

## Layout

```
src/
  api/client.js        # fetch wrapper: bearer token, list-envelope unwrap, 401 handling
  stores/auth.js       # token + user (Pinia), login/register/logout
  stores/cart.js       # multi-chef cart, persisted to localStorage
  router/index.js      # routes + auth/role guards
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

Not yet built (see issues #27–#31): favorites, reviews, search UI, real-time
chat (WebSocket), password reset screens.
