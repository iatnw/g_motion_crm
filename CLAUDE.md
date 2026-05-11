# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Layout

Monorepo with two separate runtimes:
- `backend/` — Go 1.26 API server (module: `github.com/umbrllacrp/g_motion_crm`)
- `frontend/` — Next.js 16 app (React 19, Tailwind CSS 4)

## Running the Stack

```bash
# Full stack (recommended for development)
docker-compose up

# Backend only (requires local Postgres + .env)
cd backend && air           # hot-reload via air
cd backend && go run ./cmd/server  # without hot-reload

# Frontend only
cd frontend && npm run dev
```

Backend listens on `APP_PORT` (default 8080), frontend on 3000. DB migrations run automatically on startup via `goose.Up()`.

## Backend Commands

```bash
cd backend

# Build
go build -o ./tmp/main ./cmd/server

# Test
go test ./...
go test ./internal/handler/...  # single package

# After editing SQL queries or migrations, regenerate the DB layer
sqlc generate   # reads sqlc.yaml, outputs to internal/db/
```

## Frontend Commands

```bash
cd frontend
npm run dev
npm run build
npm run lint
```

**Important:** This project uses Next.js 16 with breaking API/convention changes from earlier versions. Before writing frontend code, check `node_modules/next/dist/docs/` for the relevant guide.

## Backend Architecture

Entry point: `backend/cmd/server/main.go` wires everything together.

**Layer pattern:** `handler → service → db (sqlc-generated)`

- **`internal/config`** — loads all env vars; one config struct passed down, no global state.
- **`internal/db`** — 100% sqlc-generated from `queries/*.sql` + `migrations/`. **Never edit files here by hand.** Add/change queries in `queries/`, run `sqlc generate`.
- **`internal/service`** — all business logic. Services receive `*db.Queries` and expose domain methods. `DealService` also holds the `*pgxpool.Pool` for multi-step transactions; use `WithTx()` to get a transaction-scoped queries instance.
- **`internal/handler`** — thin HTTP adapters. Use `writeJSON` / `writeError` helpers from `response.go`. Handlers are grouped by resource and mounted in `main.go`.
- **`internal/middleware`** — JWT auth middleware; validates bearer token and checks `stage` claim (`totp_pending` vs `authenticated`).
- **`internal/storage`** — S3 wrapper for document upload, presigned URL generation, and deletion.

**HTTP router:** `go-chi/chi/v5`. Route registration in `main.go`. Rate limiting on `/auth/login` (10 req/min per IP via `go-chi/httprate`). The Tally webhook (`/webhooks/tally`) is the only unauthenticated non-auth endpoint.

**Authentication flow:**
1. `POST /auth/login` → JWT with `stage=totp_pending` or `stage=authenticated`
2. If TOTP is configured, `POST /auth/totp/verify` exchanges the pending token for an authenticated one
3. All protected routes require `stage=authenticated` in the JWT

**Encryption:** Client PII is AES-encrypted at rest. The key is a hex-encoded 32-byte value in `ENCRYPTION_KEY`. `ClientService` handles all encrypt/decrypt transparently.

**Logging:** `log/slog` with JSON handler, level INFO. All 500 responses must log via `slog.Error` before writing the response.

## Database / Migrations

Migrations live in `backend/migrations/` and run via Goose on startup. To add a migration:

```bash
cd backend
goose -dir migrations create <description> sql
```

Schema is the source of truth for sqlc — add the migration first, then write the query, then `sqlc generate`.

## Environment Variables

Copy `.env.example` to `.env`. Required vars:

| Var | Purpose |
|-----|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | HMAC secret for JWTs |
| `ENCRYPTION_KEY` | 32-byte hex key for client PII |
| `TALLY_SIGNING_SECRET` | Webhook signature validation |
| `AWS_*` / `S3_BUCKET_*` | S3 document and backup storage |
| `FRONTEND_URL` | CORS origin allowlist |
