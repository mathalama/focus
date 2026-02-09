# Telegram Bot

Minimal Telegram bot service for linking Telegram users to backend accounts.

## Run

```bash
cd telegram-bot
cp .env.example .env
set -a && source .env && set +a
go run ./cmd/bot
```

Required env vars:

- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_BOT_AUTH_TOKEN` (must match backend `TELEGRAM_BOT_AUTH_TOKEN`)

Optional:

- `BACKEND_URL` (default `http://localhost:8080`)
- `APP_URL` (default `http://localhost:5173`)
- `TELEGRAM_POLL_TIMEOUT_SECONDS` (default `30`)
- `BOT_INTERNAL_API_ADDR` (default `:8091`)

## Auth Link Flow

1. User logs into backend and requests a one-time code:

```bash
curl -X POST http://localhost:8080/api/v1/auth/telegram/link-code \
  -H "Authorization: Bearer <JWT>"
```

2. User sends to bot:

```text
/link ABCD2345
```

3. Bot calls:

- `POST /api/v1/integrations/telegram/link` with header `X-Telegram-Bot-Auth`.
- `GET /api/v1/integrations/telegram/status?telegram_user_id=<id>` with header `X-Telegram-Bot-Auth` to check if Telegram is already linked.
- `PATCH /api/v1/integrations/telegram/notifications` with header `X-Telegram-Bot-Auth` to enable/disable notifications.

## Commands

- `/start` or `/status` - check link status and show onboarding steps
- `/link <CODE>` - link Telegram using one-time code from app profile
- `/notify on` / `/notify off` - enable or disable notifications
- `/subscribe` / `/unsubscribe` - quick enable/disable notifications
- `/help` - list commands

By default, notifications are **disabled** right after linking.

The bot also shows inline buttons in each response:
- `Включить`
- `Выключить`
- `Статус`
- `Помощь`

## Internal Notify API

Bot exposes an internal API for programmatic notifications.

- `GET /health`
- `POST /internal/notify`
  - Header: `X-Telegram-Bot-Auth: <TELEGRAM_BOT_AUTH_TOKEN>`
  - Body:
    - `telegram_user_id` (required)
    - `message` (required)
    - `disable_buttons` (optional, default `false`)
    - `force` (optional, default `false`)

When `force=false`, bot checks backend link status and `notifications_enabled` first.

Example:

```bash
curl -X POST http://localhost:8091/internal/notify \
  -H "Content-Type: application/json" \
  -H "X-Telegram-Bot-Auth: <TOKEN>" \
  -d '{
    "telegram_user_id": 123456789,
    "message": "Через 5 минут начнется фокус-сессия",
    "disable_buttons": true
  }'
```
