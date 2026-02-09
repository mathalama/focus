222222# Mathalama Focus

Calm-first learning focus mode, not a productivity pressure timer.

## Stack

- Backend: Go, Gin, PostgreSQL
- Frontend: React, Vite, TypeScript, Tailwind CSS

## Product Flow (MVP)

1. Dev auth login
2. Create goal (`topic + desired result + recommended mode`)
3. Start focus session from a goal
4. Pause/resume with pause limit and interruption tracking
5. Complete session and submit mandatory reflection
6. View calm analytics overview

## Project Structure

```text
focus/
  backend/
    cmd/api/main.go
    internal/
      config/
      db/
      domain/
      http/
      repository/postgresql/
    migrations/001_init.sql
  telegram-bot/
    cmd/bot/main.go
  frontend/
    src/
      lib/api.ts
      types/index.ts
      App.tsx
```

## Quick Start

### 1) Start PostgreSQL

```bash
docker compose up -d postgres
```

### 2) Run backend

```bash
cd backend
cp .env.example .env
# optional: edit DATABASE_URL if needed
go mod tidy
go run ./cmd/api
```

Backend default: `http://localhost:8080`

### 3) Run frontend

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Frontend default: `http://localhost:5173`

## Environment

### Backend (`backend/.env`)

- `PORT=8080`
- `DATABASE_URL=postgres://mathalama:mathalama@localhost:5432/mathalama?sslmode=disable`
- `CORS_ORIGIN=http://localhost:5173`
- `MAX_SESSION_PAUSES=3`
- `TELEGRAM_BOT_AUTH_TOKEN=dev-telegram-bot-auth-change-me`
- `TELEGRAM_LINK_CODE_TTL_MINUTES=10`

### Frontend (`frontend/.env`)

- `VITE_API_BASE_URL=http://localhost:8080`

## API (MVP)

- `GET /health`
- `POST /api/v1/auth/dev-login`
- `POST /api/v1/auth/telegram/link-code` (Bearer)
- `POST /api/v1/integrations/telegram/link` (`X-Telegram-Bot-Auth`)
- `GET /api/v1/integrations/telegram/status` (`X-Telegram-Bot-Auth`, `telegram_user_id` query)
- `PATCH /api/v1/integrations/telegram/notifications` (`X-Telegram-Bot-Auth`)
- `GET /api/v1/integrations/telegram` (Bearer)
- `DELETE /api/v1/integrations/telegram` (Bearer)
- `POST /api/v1/goals`
- `GET /api/v1/goals`
- `POST /api/v1/sessions`
- `PATCH /api/v1/sessions/:sessionID/pause`
- `PATCH /api/v1/sessions/:sessionID/resume`
- `POST /api/v1/sessions/:sessionID/interruption`
- `PATCH /api/v1/sessions/:sessionID/complete`
- `POST /api/v1/sessions/:sessionID/reflection`
- `GET /api/v1/analytics/overview`

## Notes

- Auth in this scaffold is dev-mode (`/auth/dev-login`) that returns a user ID.
- Protected endpoints require header: `Authorization: Bearer <jwt_token>`.
- Reflection is mandatory in the UI after session completion.

## Roadmap: Evolution to BeeFocus

We aim to transform this MVP into a fully gamified "BeeFocus" experience.

### 1. Gamification (The Hive)
- **Nectar Currency:** Earn nectar for every minute of focused work.
- **Save the Bees:** Use nectar to collect different bee species or upgrade your virtual hive.
- **Loss Aversion:** If a session is abandoned, the gathered nectar is lost.

### 2. Audio & Atmosphere
- **Soundscapes:** Integrated background noises (Rain, Forest, Cafe, White Noise).
- **Visual Themes:** Day/Night cycle in the hive view.

### 3. Advanced Focus Tools
- **Strict Mode:** Prevent pausing or quitting once started.
- **Tagging System:** Categorize sessions by project (e.g., "Coding", "Reading") represented as colored honeycombs.
- **Smart Breaks:** Pomodoro-style automated break timers.

### 4. Community & Analytics
- **Leaderboards:** See who gathered the most nectar this week.
- **Honeycomb Heatmap:** Visual representation of daily focus consistency.
