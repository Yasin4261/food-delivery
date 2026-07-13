# Security Policy

## Supported versions

Only the latest release line receives security fixes.

| Version | Supported |
|---|---|
| latest `v4.x` tag | ✅ |
| anything older | ❌ upgrade first |

## Reporting a vulnerability

**Please do not open a public issue for security problems.**

- Preferred: [GitHub private vulnerability reporting](https://github.com/Yasin4261/food-delivery/security/advisories/new) on this repository.
- Alternatively, email the maintainer: `yasin01ysn@gmail.com` with subject `SECURITY: food-delivery`.

Include what you can: affected endpoint/component, reproduction steps, impact, and a suggested fix if you have one. You will get an acknowledgement within **72 hours**. Please give us a reasonable window to ship a fix before any public disclosure.

## Scope

In scope: the Go API (`internal/`, `cmd/`), the Vue SPA (`web/`), the deployment configuration (`docker-compose*.yml`, `deploy/`), and CI workflows.

Out of scope: denial-of-service by volume alone, reports requiring a compromised host, and vulnerabilities exclusively in third-party dependencies without a demonstrated impact here (still welcome as a heads-up).

## Security posture (summary)

The full OWASP Top 10 mapping and the **binding policies for new code** live in [CLAUDE.md §10](CLAUDE.md#10-security-policy-owasp-top-10-analysis). Highlights:

- Passwords bcrypt-hashed; reset tokens stored only as sha256, single-use, expiring; JWTs (HS256) carry a `jti` revoked on logout via a denylist (Redis-backed in multi-instance deploys).
- Role guards at the router plus **per-resource ownership checks in the service layer** (chefs own their menus, customers their orders, chat is participant-only); the `admin` role cannot be self-assigned.
- SQL exists only in `internal/repository/` and only with parameterized placeholders.
- Per-IP rate limiting on auth endpoints and the payment callback; enumeration-safe responses (forgot-password is silent for unknown emails).
- Card data never touches this codebase: payments go through iyzico hosted checkout with server-to-server verification; refunds are gateway-driven.
- Config fails fast on missing/placeholder secrets outside development; no secrets in the repo.
- Structured request logs never record credentials, tokens, or query strings.

## Keeping this document current

Any change touching authentication, authorization, payments, file handling, or new dependencies must update this file and CLAUDE.md §10 in the same pull request when the posture changes (see CONTRIBUTING.md).
