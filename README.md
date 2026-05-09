# G Motion CRM

Auto finance deal management — tracks clients, deals, brokers, dealers, referrals, and documents.

## Stack

- **Backend** — Go 1.26, chi router, PostgreSQL 16, sqlc, Goose migrations
- **Frontend** — Next.js 16, React 19, TypeScript, Tailwind CSS 4
- **Storage** — AWS S3 (documents and backups)
- **Auth** — JWT (httpOnly cookie) + optional TOTP (Google Authenticator)
- **Webhooks** — Tally form submissions auto-create clients and deals

## Local Development

```bash
# 1. Copy and fill in environment variables
cp .env.example .env

# 2. Start everything (Postgres + backend with hot-reload + frontend)
docker-compose up
```

- Backend: http://localhost:8080
- Frontend: http://localhost:3000
- DB migrations run automatically on backend startup

## Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret for signing JWTs (any long random string) |
| `ENCRYPTION_KEY` | 32-byte hex key for encrypting client PII — generate with `openssl rand -hex 32` |
| `TALLY_SIGNING_SECRET` | Found in your Tally webhook settings |
| `AWS_REGION` | AWS region for S3 (e.g. `us-east-1`) |
| `AWS_ACCESS_KEY_ID` | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key |
| `S3_BUCKET_DOCUMENTS` | S3 bucket name for uploaded documents |
| `S3_BUCKET_BACKUPS` | S3 bucket name for DB backups |
| `FRONTEND_URL` | CORS allowed origin (e.g. `https://yourapp.com`) |
| `APP_ENV` | Set to `production` to enable Secure flag on cookies |
| `TOTP_ISSUER` | Name shown in the authenticator app (e.g. `GMotionCRM`) |

## Database Migrations

Migrations live in `backend/migrations/` and run automatically on startup via Goose. To create a new migration:

```bash
cd backend
goose -dir migrations create <description> sql
```

After adding a migration, update the DB layer:

```bash
cd backend
sqlc generate
```

## Running Tests

```bash
cd backend

# Unit tests (no database needed)
go test ./internal/handler/... ./internal/service/...

# Integration tests (requires a running Postgres)
TEST_DATABASE_URL="postgres://user:pass@localhost:5432/gmotion_test?sslmode=disable" \
  go test ./internal/integration/... -timeout 60s
```

## Production Deployment

Use the production Dockerfiles:

```bash
# Backend
docker build -f backend/Dockerfile.prod -t gmotion-backend ./backend

# Frontend
docker build -f frontend/Dockerfile.prod -t gmotion-frontend ./frontend
```

Set `APP_ENV=production` in your environment so cookies get the `Secure` flag.

## Tally Webhook

Configure your Tally form to POST to `POST /webhooks/tally`. The endpoint:

- Verifies the HMAC-SHA256 signature using `TALLY_SIGNING_SECRET`
- Creates or updates a client from the form fields
- Creates a deal linked to that client
- Downloads and stores any uploaded files (driver's license, insurance card) to S3

Expected field labels in the Tally form: `Full legal name`, `Phone Number`, `Email`, `Vehicle of Interest`, `Social Security Number`, plus address, employer, and income fields.

## CI

GitHub Actions runs on every push and pull request to `main`:
- Backend: `go build`, `go vet`, unit tests, integration tests (with a real Postgres)
- Frontend: `npm ci`, ESLint, `next build`
