# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project Overview

Monorepo for an application management platform: **Go backend** (`app-manager/`) + **React frontend** (`manager-frontend-app/`). Manages Docker containers and system processes on Ubuntu servers via a web dashboard. The UI language is Chinese (Simplified).

## Development Commands

### One-command dev (Windows)
```powershell
cd manager-frontend-app; npm install; cd ..
npm run dev        # starts both backend (:8080) and frontend (:5173) as PowerShell jobs
npm run stop       # kills the background jobs
```

### Backend (Go)
```bash
cd app-manager
go run ./cmd/app-manager              # run dev server
go build -o app-manager ./cmd/app-manager  # build binary
go test ./...                         # run all tests
go test ./internal/config/...         # run a single package's tests
```

### Frontend (React + Vite)
```bash
cd manager-frontend-app
npm install
npm run dev        # Vite dev server at :5173, proxies /api → localhost:8080
npm run build      # tsc + vite build → dist/
npm run lint       # eslint
npm run preview    # preview production build
```

## Backend Architecture

Go module: `github.com/Skyinfi/management-platform/app-manager` (Go 1.22, no external HTTP framework).

**Layered design:** `main.go` → `app` → `httpapi` → `service` → `store`

- **`cmd/app-manager/main.go`** — entry point: loads config, wires store/auth/server, starts `http.ListenAndServe`
- **`internal/app/app.go`** — thin wiring layer (factory functions)
- **`internal/httpapi/server.go`** — HTTP handlers, route registration via `http.ServeMux`, functional options (`WithJWTValidator`, `WithCORS`)
- **`internal/httpapi/auth.go`** — login/me handlers
- **`internal/service/service.go`** — business logic (dashboard, applications, logs, actions)
- **`internal/service/auth.go`** — authentication logic (login, me)
- **`internal/store/store.go`** — in-memory data store with `sync.RWMutex` (currently seeded with mock data)
- **`internal/model/`** — request/response structs
- **`internal/middleware/`** — `Recovery`, `Logging`, `CORS`, `Auth` middleware chain; custom HMAC-SHA256 JWT implementation (not a library)
- **`internal/config/config.go`** — env-var-driven config with defaults

**API routes** (all prefixed `/api/`):
- `GET /health`, `POST /auth/login`, `GET /auth/me` — public (no auth required)
- `GET /dashboard`, `GET /applications`, `GET /applications/{name}/logs`, `POST /applications/{name}/{action}` — auth-protected

**Auth flow:** Login returns a custom JWT (base64url payload + HMAC-SHA256 signature). All routes except health/login/me require `Authorization: Bearer <token>`.

## Frontend Architecture

React 19 + TypeScript + Vite. No router library — single-page layout with conditional rendering (login vs dashboard).

- **`src/api/client.ts`** — fetch-based HTTP client with auth token injection and 401 auto-logout
- **`src/api/`** — typed API modules (`appManager.ts`, `auth.ts`, `types.ts`)
- **`src/auth/`** — `useAuth` hook + `authStore` (localStorage persistence, bootstrap on mount)
- **`src/hooks/`** — data hooks (`useDashboard`, `useApplications`, `useAppActions`, `useAppLogs`) that fall back to mock data
- **`src/components/`** — UI components (sidebar, tables, cards, log panel, login page)
- **`src/types.ts`** — shared TypeScript interfaces

Vite dev server proxies `/api` to `http://localhost:8080` (configured in `vite.config.ts`).

## Backend Environment Variables

| Variable | Default |
|----------|---------|
| `APP_MANAGER_ADDR` | `:8080` |
| `APP_MANAGER_ENV` | `development` |
| `APP_MANAGER_LOG_LEVEL` | `info` |
| `APP_MANAGER_ENABLE_CORS` | `true` |
| `APP_MANAGER_JWT_SECRET` | `dev-secret` |
| `APP_MANAGER_ALLOW_ORIGIN` | `*` |

## Key Conventions

- All API responses use `{ code, message, data }` envelope (`model.APIResponse`)
- Backend uses `net/http` standard library only (no Gin/Echo) — handlers are methods on `httpapi.Server`
- Frontend API base URL configurable via `VITE_API_BASE_URL` env var, defaults to `/api`
- Auth token stored in `localStorage` under keys prefixed `app-manager-`