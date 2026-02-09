# Mathalama Focus

Calm-first focus and reflection app with web UI, API, and Telegram bot integration.

## Stack

- Backend: Go, Gin, PostgreSQL
- Telegram bot: Go (long polling + internal notify API)
- Frontend: React, Vite, TypeScript, Tailwind CSS
- Infra: Docker Compose

## Project Structure

```text
focus/
  backend/
    cmd/
      api/
      migrate/
    internal/
      config/
      delivery/httpapi/
      domain/
      infrastructure/
      usecase/
    migrations/
    Dockerfile
  telegram-bot/
    cmd/bot/
    internal/
    Dockerfile
  frontend/
    src/
      api/
      components/
      context/
      pages/
```

## Quick Start (Docker + npm)

1. Prepare env files:

```bash
cp backend/.env.example backend/.env
cp telegram-bot/.env.example telegram-bot/.env
cp frontend/.env.example frontend/.env
```

2. Fill required secrets:

- `backend/.env`: `JWT_SECRET`, `TELEGRAM_BOT_AUTH_TOKEN`
- `telegram-bot/.env`: `TELEGRAM_BOT_TOKEN`, same `TELEGRAM_BOT_AUTH_TOKEN` as backend

3. Start database, migrations, backend, bot:

```bash
docker compose up -d --build postgres migrate backend telegram-bot
```

4. Start frontend without Docker:

```bash
cd frontend
npm install
npm run dev
```

Endpoints:

- Frontend: `http://localhost:5173`
- Backend health: `http://localhost:8080/health`
- Telegram bot internal API health: `http://localhost:8091/health`

## Local Backend/Bot Without Docker

Run migrations first:

```bash
cd backend
go run ./cmd/migrate
go run ./cmd/api
```

Then run bot:

```bash
cd telegram-bot
go run ./cmd/bot
```

## Migrations

- SQL files are stored in `backend/migrations`.
- Migration runner: `backend/cmd/migrate`.
- Applied migrations are tracked in `schema_migrations` with checksum validation.
- Docker Compose uses a dedicated `migrate` service that runs before `backend`.

## Environment

Backend (`backend/.env`):

- `PORT=8080`
- `DATABASE_URL=postgres://mathalama:mathalama@localhost:5432/mathalama?sslmode=disable`
- `CORS_ORIGIN=http://localhost:5173`
- `JWT_SECRET=dev-secret-change-me`
- `ENABLE_DEV_LOGIN=true`
- `MAX_SESSION_PAUSES=3`
- `TELEGRAM_BOT_AUTH_TOKEN=dev-telegram-bot-auth-change-me`
- `TELEGRAM_LINK_CODE_TTL_MINUTES=10`
- `RESEND_API_KEY=`
- `RESEND_FROM_EMAIL=`
- `EMAIL_VERIFY_URL_BASE=http://localhost:8080/api/v1/auth/verify-email`
- `EMAIL_VERIFY_SUCCESS_REDIRECT=http://localhost:5173/login?verified=1`
- `EMAIL_VERIFY_FAIL_REDIRECT=http://localhost:5173/login?verified=0`
- `EMAIL_VERIFICATION_TTL_MINUTES=60`

Telegram bot (`telegram-bot/.env`):

- `TELEGRAM_BOT_TOKEN=`
- `TELEGRAM_BOT_AUTH_TOKEN=dev-telegram-bot-auth-change-me`
- `BACKEND_URL=http://localhost:8080`
- `APP_URL=http://localhost:5173`
- `TELEGRAM_POLL_TIMEOUT_SECONDS=30`
- `BOT_INTERNAL_API_ADDR=:8091`

Frontend (`frontend/.env`):

- `VITE_API_BASE_URL=http://localhost:8080`

## Security Notes

- `ENABLE_DEV_LOGIN` should be `false` outside local development.
- Compose sets `ENABLE_DEV_LOGIN=false` for the `backend` service.
- Keep `.env` files local; they are ignored by git.
