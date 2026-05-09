# G Motion CRM — Frontend

Next.js 16 frontend for the G Motion CRM. See the [root README](../README.md) for setup, environment variables, and deployment instructions.

## Development

```bash
npm run dev    # dev server on http://localhost:3000
npm run build  # production build
npm run lint   # ESLint
```

The frontend expects the backend running on `NEXT_PUBLIC_API_URL` (defaults to `http://localhost:8080`). Use `docker-compose up` from the repo root to start the full stack.
