# Copilot / AI agent instructions — Mathalama Focus

Purpose: make an AI coding agent immediately productive in this repo by summarizing architecture, workflows, conventions, and exact edit points.

- **Big picture**: Backend is a Go HTTP API (Gin) serving the app data and gamification logic; frontend is a Vite + React + TypeScript SPA. Key entrypoints:
  - Backend: backend/cmd/api/main.go
  - Router: backend/internal/http/router.go
  - Handlers: backend/internal/http/handlers/handlers.go (use the `Repository` interface)
  - Repository / persistence: backend/internal/repository/postgresql/repository.go
  - DB pool: backend/internal/db/postgres.go
  - Config: backend/internal/config/config.go (reads env via godotenv)
  - Frontend: frontend/src (API wrapper at frontend/src/lib/api.ts, auth in frontend/src/context/AuthContext.tsx)

- **How to run locally (explicit)**:
  1. Start Postgres (docker compose mounts backend/migrations into initdb):

     docker compose up -d postgres

  2. Backend (expects `DATABASE_URL` env):

     cd backend
     cp .env.example .env
     go mod tidy
     go run ./cmd/api

     Backend defaults: http://localhost:8080 (configured in `backend/internal/config/config.go`).

  3. Frontend:

     cd frontend
     cp .env.example .env
     npm install
     npm run dev

     Frontend defaults: Vite at http://localhost:5173; `VITE_API_BASE_URL` controls API host.

- **API & auth conventions (project-specific)**:
  - Dev auth endpoint: POST /api/v1/auth/dev-login — returns `{ user, token }` using `auth.GenerateToken`.
  - Protected routes use `middleware.RequireUserID(jwtSecret)` in `router.go` and expect `Authorization: Bearer <jwt>`.
  - All API shapes are implemented in `frontend/src/lib/api.ts` — mirror any server changes there.

- **Persistence & migrations**:
  - Migrations are in `backend/migrations` and are mounted into the Postgres container via `docker-compose.yml` for automatic init.
  - Repository methods return domain types in `backend/internal/domain` and surface sentinel errors in `repository/postgresql/repository.go` (e.g. `ErrNotFound`, `ErrPauseLimitReached`, `ErrInvalidState`). Handlers convert those to HTTP responses.

- **Patterns to follow when editing**:
  - Handlers depend on a `Repository` interface (see `handlers.go`). For new endpoints:
    1. Add method to `Repository` interface in `handlers.go`.
    2. Implement it in `repository/postgresql/repository.go` (return domain types and use sentinel errors where appropriate).
    3. Wire handler in `backend/internal/http/router.go`.
  - When changing DB schemas: add a migration to `backend/migrations` and bump logic in repository code. Tests or manual migration runs are expected.

- **Developer workflows & checks**:
  - DB pool settings are in `backend/internal/db/postgres.go` (MaxConns=10, MinConns=1) — consider these when writing integration tests.
  - Use `go vet`, `gofmt`, and `go test ./...` for backend; `npm run typecheck` and `npm run build` for frontend checks.

- **Integration examples (use these exact endpoints/headers in tests or stub flows)**:
  - Dev-login (curl):

    curl -X POST http://localhost:8080/api/v1/auth/dev-login -H 'Content-Type: application/json' -d '{"email":"dev@example.com","name":"Dev"}'

  - Create goal (requires Bearer token): POST /api/v1/goals with JSON body `{ topic, desired_result, recommended_minutes, tags }`.

- **Files to inspect first for any change request**:
  - backend/cmd/api/main.go
  - backend/internal/http/router.go
  - backend/internal/http/handlers/handlers.go
  - backend/internal/repository/postgresql/repository.go
  - backend/internal/config/config.go
  - frontend/src/lib/api.ts

If anything here is unclear or you want additional examples (test harness, example migrations, or a local dev script), tell me which area to expand. 
