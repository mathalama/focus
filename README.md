# Mathalama Focus

> Calm-first focus and reflection app with web UI, API, and Telegram bot integration.

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=flat&logo=typescript)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

---

## → Technology Stack

| Component | Technology |
|-----------|-----------|
| **★ Backend** | Go · Gin · PostgreSQL |
| **★ Telegram Bot** | Go · Long Polling · REST API |
| **★ Frontend** | React · Vite · TypeScript · Tailwind CSS |
| **★ Infrastructure** | Docker Compose · PostgreSQL |

---

## → Project Structure

```
focus/
├── backend/                  ★ REST API & Core Business Logic
│   ├── cmd/
│   │   ├── api/              ★ API Server
│   │   └── migrate/          ★ Database Migration Runner
│   ├── internal/
│   │   ├── config/           ★ Configuration Management
│   │   ├── delivery/httpapi/ ★ HTTP Endpoints
│   │   ├── domain/           ★ Domain Models & Errors
│   │   ├── infrastructure/   ★ External Services (Email, JWT, DB)
│   │   └── usecase/          ★ Business Logic
│   ├── migrations/           ★ SQL Migration Files
│   └── Dockerfile
│
├── telegram-bot/            ★ Telegram Bot Service
│   ├── cmd/bot/
│   ├── internal/
│   └── Dockerfile
│
├── frontend/                ★ Web Application (React)
│   ├── src/
│   │   ├── api/             ★ API Client
│   │   ├── components/      ★ UI Components
│   │   ├── context/         ★ React Context
│   │   └── pages/           ★ Page Components
│   ├── public/
│   └── vite.config.ts
│
└── browser-extension/       ★ Browser Extension (TypeScript)
    ├── src/
    │   ├── background/
    │   ├── content/
    │   └── popup/
    └── manifest.json
```

---

## → Quick Start with Docker

### ► Step 1: Prepare Environment Files

```bash
cp backend/.env.example backend/.env
cp telegram-bot/.env.example telegram-bot/.env
cp frontend/.env.example frontend/.env
```

### ► Step 2: Configure Secrets

Edit environment files with required values:

**`backend/.env`**
```
JWT_SECRET=your_jwt_secret_here
TELEGRAM_BOT_AUTH_TOKEN=your_bot_auth_token
```

**`telegram-bot/.env`**
```
TELEGRAM_BOT_TOKEN=your_telegram_token
TELEGRAM_BOT_AUTH_TOKEN=your_bot_auth_token  # Same as backend
```

### ► Step 3: Start Services

```bash
docker compose up -d --build postgres migrate backend telegram-bot
```

### ► Step 4: Start Frontend

```bash
cd frontend
npm install
npm run dev
```

### ► Service Endpoints

| Service | URL | Purpose |
|---------|-----|---------|
| **Frontend** | `http://localhost:5173` | Web Application |
| **Backend Health** | `http://localhost:8080/health` | API Health Check |
| **Backend Metrics** | `http://localhost:8080/metrics` | Prometheus Metrics |
| **Bot Health** | `http://localhost:8091/health` | Telegram Bot API |

---

## → Local Development (No Docker)

### ► Backend & Migrations

```bash
cd backend
go run ./cmd/migrate  # Run database migrations first
go run ./cmd/api      # Start API server
```

### ► Telegram Bot

```bash
cd telegram-bot
go run ./cmd/bot
```

### ► Frontend

```bash
cd frontend
npm install
npm run dev
```

---

## → Database Migrations

**► Location:** `backend/migrations/`

**► How It Works:**
• SQL files are versioned and tracked in `schema_migrations` table
• Migration runner: `backend/cmd/migrate`
• Checksum validation prevents duplicate/corrupted migrations
• Docker Compose runs migrations automatically via dedicated service

**► Available Migrations:**
```
001_init.sql
002_add_nectar.sql
003_roadmap_features.sql
004_telegram_auth.sql
005_telegram_notifications.sql
006_email_auth.sql
007_auth_sessions_rbac.sql
008_email_outbox.sql
009_product_events.sql
010_password_reset.sql
011_notification_sounds.sql
012_user_session_preferences.sql
013_notification_schedules.sql
014_email_outbox_email_type_and_reset_link.sql
```

---

## → Environment Configuration

### ► Backend (`backend/.env`)

**○ Server:**
• `PORT=8080`
• `DATABASE_URL=postgres://mathalama:mathalama@localhost:5432/mathalama?sslmode=disable`
• `CORS_ORIGIN=http://localhost:5173`

**○ Authentication:**
• `JWT_SECRET=dev-secret-change-me`
• `ENABLE_DEV_LOGIN=true`
• `REFRESH_SESSION_TTL_HOURS=720`
• `MAX_ACTIVE_AUTH_SESSIONS=5`
• `AUTH_SESSION_BIND_CLIENT=true`
• `MAX_SESSION_PAUSES=3`

**○ Integrations:**
• `TELEGRAM_BOT_AUTH_TOKEN=dev-telegram-bot-auth-change-me`
• `TELEGRAM_LINK_CODE_TTL_MINUTES=10`
• `RESEND_API_KEY=`
• `RESEND_FROM_EMAIL=`

**○ Email Verification:**
• `EMAIL_OUTBOX_POLL_SECONDS=2`
• `EMAIL_OUTBOX_MAX_ATTEMPTS=5`
• `EMAIL_VERIFY_URL_BASE=http://localhost:8080/api/v1/auth/verify-email`
• `EMAIL_VERIFY_SUCCESS_REDIRECT=http://localhost:5173/login?verified=1`
• `EMAIL_VERIFY_FAIL_REDIRECT=http://localhost:5173/login?verified=0`
• `EMAIL_VERIFICATION_TTL_MINUTES=60`

### ► Telegram Bot (`telegram-bot/.env`)

**○ API Configuration:**
• `TELEGRAM_BOT_TOKEN=`
• `TELEGRAM_BOT_AUTH_TOKEN=dev-telegram-bot-auth-change-me`
• `BACKEND_URL=http://localhost:8080`
• `APP_URL=http://localhost:5173`
• `TELEGRAM_POLL_TIMEOUT_SECONDS=30`
• `BOT_INTERNAL_API_ADDR=:8091`

### ► Frontend (`frontend/.env`)

**○ API:**
• `VITE_API_BASE_URL=http://localhost:8080`

---

## → Security & Authentication

### ► Best Practices

**► Development Only:**
• `ENABLE_DEV_LOGIN` must be `false` in production
• Docker Compose already sets `ENABLE_DEV_LOGIN=false` for backend service
• `.env` files are gitignored — keep them local

**► Idempotency:**
Include `Idempotency-Key` header on critical write operations:
• `PATCH /api/v1/sessions/:sessionID/complete`
• `POST /api/v1/shop/items/:itemID/buy`
• `POST /api/v1/integrations/telegram/link`

### ► Authentication Flow (v2)

**Login Response:**
```json
{
  "token": "access_jwt",
  "refresh_token": "refresh_jwt"
}
```

**Available Endpoints:**
• `POST /api/v1/auth/refresh` — Refresh access token
• `POST /api/v1/auth/logout` — Logout current session
• `POST /api/v1/auth/logout-all` — Logout all sessions (Bearer token required)
• `GET /api/v1/auth/sessions` — List active sessions (Bearer token required)

---

## → Admin API

Requires Bearer token authentication and admin role:

• `GET /api/v1/admin/health` — Health check
• `GET /api/v1/admin/email-deliveries` — Email delivery status
• `GET /api/v1/admin/events` — Product events

---

## → Product Observability

**► Request Tracking:**
• Structured JSON logs with `request_id` (via `X-Request-ID` header)
• Unique trace ID for every API request

**► Metrics:**
• Prometheus-compatible metrics available at `/metrics`
• Response times, request counts, error rates

**► Event System:**
• Product events persisted in `product_events` table
• Email verification via async DB outbox (`email_outbox`)
• Automatic retry with exponential backoff on delivery failure

---

## → Branching & Release Strategy

**► Branch Structure:**

| Branch | Purpose | Phase Gate |
|--------|---------|-----------|
| `dev` | Internal testing | internal-testing |
| `stage` | Open testing | open-testing |
| `prod` | Production release | release |

**► CI/CD Pipeline:**
• Same build validation runs on all branches
• Branch-aware deployment triggers appropriate phase gates
• Supports canary deployments and gradual rollouts

---

## → Contributing

**► Getting Started:**
1. Fork the repository
2. Create feature branch from `dev`
3. Submit pull request to `dev` first
4. After testing, PR can be merged to `stage` then `prod`

**► Code Style:**
• Backend: Go conventions (gofmt)
• Frontend: ESLint + Prettier
• Migrations: Numbered SQL files with descriptive names

---
