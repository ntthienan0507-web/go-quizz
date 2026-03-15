# Go Quizz

Real-time quiz platform with live multiplayer, WebSocket game sessions, leaderboards & instant scoring.

> Backend architecture follows patterns from [create-go-api](https://github.com/ntthienan0507-web/create-go-api) — 3-layer architecture, repository interfaces, dependency injection, adapter pattern.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25+, Gin, GORM, gorilla/websocket |
| Frontend | React 19, TypeScript, React Router v7 |
| Database | PostgreSQL 16 |
| Cache / Real-time | Redis 7 (sorted sets for leaderboard) |
| Auth | JWT + refresh token rotation |
| Infra | Docker Compose, GitHub Actions CI, Render (BE), Netlify (FE) |

## Architecture

```
┌─────────────┐     WebSocket      ┌──────────────────────────────────┐
│   React UI  │◄──────────────────►│  WebSocket Hub (per quiz room)   │
│  (browser)  │     REST API       │                                  │
│             │◄──────────────────►│         Gin HTTP Server          │
└─────────────┘                    └──────┬──────────┬────────────────┘
                                          │          │
                                   ┌──────▼───┐ ┌────▼─────┐
                                   │PostgreSQL│ │  Redis   │
                                   │  (GORM)  │ │ (scores, │
                                   │          │ │  leaderb)│
                                   └──────────┘ └──────────┘
```

**Backend follows 3-layer architecture:**

```
Controller  →  Service  →  Repository
(HTTP/WS)     (logic)      (data access)
```

Each module (`auth`, `quiz`, `player`) is self-contained with its own controller, service, repository, routes, models, and types. Repositories are interfaces — implementations can be swapped without touching business logic.

See [backend/ARCHITECTURE.md](./backend/ARCHITECTURE.md) for all 12 design patterns used.

## Features

- **Live Quiz Mode** — Host starts quiz, controls question flow, all players answer in real-time
- **Self-Paced Mode** — Players go through questions at their own speed
- **Guest Play** — Join with just a name, no account required
- **Real-time Leaderboard** — Redis sorted sets, broadcast via WebSocket
- **Instant Scoring** — Points based on correctness + speed
- **JWT Auth** — 15-min access token, 7-day single-use refresh token
- **Rate Limiting** — Redis sliding window per-IP and per-user
- **Auto Seed** — Sample users + quizzes created on first run

## Quick Start

### 1. Start infrastructure

```bash
docker compose up postgres redis -d
```

> Ports: PostgreSQL → `5433`, Redis → `6380` on host.

### 2. Backend

```bash
cd backend
cp .env.example .env    # edit if needed
go mod download
make dev                # hot-reload with Air
```

Server: http://localhost:8080 — Swagger: http://localhost:8080/swagger/index.html

### 3. Frontend

```bash
cd frontend
npm install
npm start
```

App: http://localhost:3000

### Or run everything with Docker

```bash
docker compose up -d
```

| Service | URL |
|---------|-----|
| Frontend | http://localhost:5175 |
| Backend | http://localhost:8080 |
| PostgreSQL | localhost:5433 |
| Redis | localhost:6380 |

## Seed Data

Auto-created on first run:

| Email | Password | Role |
|-------|----------|------|
| `admin@test.com` | `123456` | admin |
| `player1@test.com` | `123456` | player |
| `player2@test.com` | `123456` | player |

Sample quizzes: `GENKNOW` (live), `SCITECH` (self-paced).

## Project Structure

```
go-quizz/
├── backend/
│   ├── main.go                    # Entry point, DI wiring
│   ├── modules/
│   │   ├── auth/                  # JWT login, register, refresh
│   │   ├── quiz/                  # CRUD, start/finish lifecycle
│   │   │   ├── question/          # Question management (nested)
│   │   │   └── realtime/          # Redis scoring + leaderboard
│   │   └── player/                # Dashboard, history, profile
│   ├── pkg/
│   │   ├── config/                # Env config loader
│   │   ├── db/                    # PostgreSQL + Redis + seed
│   │   ├── middleware/            # Auth, CORS, rate limiting
│   │   ├── response/              # Standard JSON envelope
│   │   └── ws/                    # WebSocket hub, client, handler
│   ├── ARCHITECTURE.md            # 12 design patterns
│   └── README.md                  # Backend setup & troubleshooting
├── frontend/
│   ├── src/
│   │   ├── pages/                 # Route pages (dashboard, quiz, join...)
│   │   ├── components/            # Shared components
│   │   ├── context/               # Auth context
│   │   ├── hooks/                 # useWebSocket
│   │   ├── api/                   # Axios HTTP client
│   │   └── ws/                    # WebSocket client
│   └── e2e/                       # Puppeteer E2E tests
├── docker-compose.yml
└── .github/workflows/ci.yml       # Build + test + lint
```

## Design Patterns (Backend)

| # | Pattern | Purpose | File |
|---|---------|---------|------|
| 1 | Three-Layer Architecture | Controller → Service → Repository | `modules/*/` |
| 2 | Repository Interface | Swap DB impl without touching logic | `modules/*/repository.go` |
| 3 | Dependency Injection | Constructor injection, testable | `main.go` |
| 4 | Adapter Pattern | Bridge realtime services to WS interfaces | `modules/quiz/realtime/adapter.go` |
| 5 | Hub Pattern | Per-quiz WebSocket room with broadcast | `pkg/ws/hub.go` |
| 6 | Dual Auth Strategy | JWT token + guest query param for WS | `pkg/ws/handler.go` |
| 7 | JWT + Refresh Rotation | 15-min access, 7-day single-use refresh | `modules/auth/service.go` |
| 8 | Rate Limiting | Redis sliding window, fail-open | `pkg/middleware/ratelimit.go` |
| 9 | Middleware Chain | Recovery → RequestID → CORS → Auth → Handler | `pkg/middleware/` |
| 10 | Redis Real-Time Store | Sorted sets for leaderboard during game | `modules/quiz/realtime/redis.go` |
| 11 | Quiz Lifecycle FSM | draft → active → finished with side effects | `modules/quiz/service.go` |
| 12 | Response Envelope | `{success, data, error, meta}` | `pkg/response/response.go` |

Full documentation: [backend/ARCHITECTURE.md](./backend/ARCHITECTURE.md)

## Scripts

```bash
# Backend
make build          # Build binary
make run            # go run
make dev            # Air hot-reload
make test           # go test -race
make lint           # golangci-lint

# Frontend
npm start           # Dev server (port 3000)
npm run build       # Production build
npm test            # Jest tests
```

## Deployment

### Backend — [Render](https://render.com)

Uses `render.yaml` blueprint (free tier):
- **Web Service**: Go backend (Docker)
- **PostgreSQL**: Managed database
- **Redis**: Managed cache

Deploy: Render Dashboard → New → Blueprint → connect repo.

Environment variables are auto-wired via `render.yaml`. Backend supports both `DATABASE_URL`/`REDIS_URL` (Render) and individual `DB_HOST`/`REDIS_ADDR` (local).

### Frontend — [Netlify](https://netlify.com)

Deploy: Netlify → New site from Git → set:
- **Base directory**: `frontend`
- **Build command**: `npm run build`
- **Publish directory**: `frontend/build`
- **Env vars**:
  - `REACT_APP_API_URL` = `https://<your-render-app>.onrender.com/api`
  - `REACT_APP_WS_URL` = `wss://<your-render-app>.onrender.com/ws`

## CI/CD

GitHub Actions runs on every push/PR to `main`:
- **Backend** — build, test (`-race`), lint (`golangci-lint`)
- **Frontend** — install, type-check (`tsc --noEmit`), build

## License

Private project.
