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

# 2. Build and start Postgres, the Go watcher, and the Next.js dev server
docker-compose up --build
```

- Backend: http://localhost:8080
- Frontend: http://localhost:3000
- DB migrations run automatically on backend startup

Backend and frontend sources are mounted into their development containers. Air
rebuilds and restarts the API after Go or migration changes; Next.js refreshes the
browser after frontend changes. Both use polling for Docker Desktop compatibility.
The backend runs with `APP_ENV=development` so local HTTP cookies work.

Container dependencies and build output are isolated from your host's
`node_modules`, `.next`, and backend temporary files. After changing frontend
`package.json` or `package-lock.json`, rebuild and refresh its dependency volume:

```bash
docker-compose up --build --renew-anon-volumes frontend
```

Rebuild the backend image after changing `go.mod` or `go.sum`. Stop development
containers with `docker-compose down`; this preserves the database volume.

## Local sample data and storage

Development Compose includes a local S3 simulator at `http://localhost:9000`,
with persistent document and backup volumes. Set `LOCAL_S3_SECRET_KEY` in `.env`
(generate it with `openssl rand -hex 32`) before starting Compose. The development
containers use local credentials and buckets; no AWS account is needed.

After `docker-compose up --build -d`, populate an empty database and add the
comprehensive workflow examples:

```bash
# Run once on an empty database (this original seed is not repeatable).
docker-compose exec backend go run ./cmd/seeddata
# Safe to repeat: preserves existing data and skips previously added scenarios.
docker-compose exec backend go run ./cmd/seedscenarios
```

The scenario set covers cash deals (paid/pending/cancelled), new leads without
VINs, cancelled financing, active/expiring/ended leases, cosigners and their
cancelled intake, repeat clients, overdue balances and reminder flags, fixed and
percentage commissions, all broker payment methods, zero-fee and direct deals,
full and incomplete client profiles, all client sources/employment statuses,
linked and direct referrals in every status, and empty broker/dealer statistics.
Dates are relative to the seeding date; choose a wider dashboard date range to
see historical leases. New records are marked as fictional demo data.

Seven clearly labelled text placeholders cover every document category and can
be downloaded normally. A synthetic Tally intake also exercises submission
processing and duplicate detection without contacting Tally. It does not simulate
external file downloads. Passwords and TOTP settings are preserved.

The local storage service uses [Adobe S3Mock](https://github.com/adobe/S3Mock),
a development simulator that does not validate S3 signatures. It is exposed only
on loopback. Production Compose continues to use AWS S3. `S3_ENDPOINT` selects an
optional internal S3-compatible endpoint, and `S3_PUBLIC_ENDPOINT` controls the
browser-accessible host in presigned URLs; leave both empty for AWS.

## Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret for signing JWTs (any long random string) |
| `ENCRYPTION_KEY` | 32-byte hex key for encrypting client PII — generate with `openssl rand -hex 32` |
| `SETUP_SECRET` | Separate random secret required to claim an uninitialized deployment — generate with `openssl rand -hex 32` |
| `TALLY_SIGNING_SECRET` | Found in your Tally webhook settings |
| `AWS_REGION` | AWS region for S3 (e.g. `us-east-1`) |
| `AWS_ACCESS_KEY_ID` | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key |
| `S3_BUCKET_DOCUMENTS` | S3 bucket name for uploaded documents |
| `S3_BUCKET_BACKUPS` | S3 bucket name for DB backups |
| `FRONTEND_URL` | CORS allowed origin (e.g. `https://yourapp.com`) |
| `APP_ENV` | Set to `production` to enable Secure flag on cookies |
| `TOTP_ISSUER` | Name shown in the authenticator app (e.g. `GMotionCRM`) |

Disabling TOTP requires the current password and an authenticator code.

Password changes and logout invalidate all owner sessions on all devices. Existing JWTs without a session version require a fresh login after upgrading.

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

`docker-compose.prod.yml` is a standalone production file — no local Postgres (uses AWS RDS instead), no source mounts, multi-stage builds, `restart: unless-stopped`.

**Important:** Production on EC2 must be deployed with `docker-compose.prod.yml`.
The default `docker-compose.yml` is for local/dev-style runs because it starts a
local Postgres container. Do not use `docker-compose up` for production deploys.

**Infrastructure needed:**
- EC2 instance (t3.small recommended) with Docker installed
- RDS PostgreSQL instance — set `DATABASE_URL` to the RDS endpoint in `.env`
- Two S3 buckets (documents + backups)

**First deployment on EC2:**
```bash
git clone git@github.com:umbrllacrp/g_motion_crm.git
cd g_motion_crm
cp .env.example .env
# Fill in all values — DATABASE_URL points to your RDS endpoint,
# APP_ENV=production, BACKEND_URL=https://api.yourcrm.com
nano .env

docker-compose -f docker-compose.prod.yml up -d --build
```

**Deploying a fix:**
```bash
git pull
docker-compose -f docker-compose.prod.yml up -d --build
```

**Verifying production is using RDS:**
```bash
docker exec g_motion_crm-backend-1 sh -lc \
  'printenv DATABASE_URL | sed -E "s#postgres://[^@]+@([^/:?]+).*#DATABASE_URL_HOST=\1#"'
```

Expected output should be the RDS host, not `postgres`.

**Viewing logs:**
```bash
docker logs g_motion_crm-backend-1 -f
docker logs g_motion_crm-frontend-1 -f
```

Set `APP_ENV=production` so cookies get the `Secure` flag.
Set `BACKEND_URL` to the public URL the browser uses to reach the backend (e.g. `https://api.yourcrm.com`).

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
