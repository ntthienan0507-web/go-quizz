# Quiz Platform — Backend

Real-time quiz platform API with WebSocket-based live games, leaderboards, and scoring.

## Tech Stack

- **Go 1.25+** — Language
- **Gin** — HTTP framework
- **GORM** — ORM (PostgreSQL)
- **go-redis** — Redis client (leaderboard, scoring, rate limiting)
- **gorilla/websocket** — Real-time quiz sessions
- **golang-jwt** — JWT authentication
- **Air** — Hot-reload for development

## Prerequisites

| Dependency | Required Version | Check Command |
|-----------|-----------------|---------------|
| Go | 1.25+ | `go version` |
| PostgreSQL | 14+ | `psql --version` |
| Redis | 6+ | `redis-cli --version` |
| Air (optional) | latest | `air -v` |
| golangci-lint (optional) | v2+ | `golangci-lint version` |

## Quick Start

### 1. Start services (PostgreSQL + Redis)

**Option A — Docker Compose** (recommended):

```bash
# From project root (one level up)
docker compose up postgres redis -d
```

> Note: docker-compose maps **postgres → port 5433** and **redis → port 6380** on host.

**Option B — Local services:**

```bash
# macOS
brew install postgresql@16 redis
brew services start postgresql@16
brew services start redis

# Ubuntu/Debian
sudo apt install postgresql redis-server
sudo systemctl start postgresql redis
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your values
```

### 3. Create database (if running locally)

```bash
createdb -U postgres quizdb
# Or with custom user:
psql -U postgres -c "CREATE USER quizuser WITH PASSWORD 'quizpass';"
psql -U postgres -c "CREATE DATABASE quizdb OWNER quizuser;"
```

### 4. Install dependencies & run

```bash
go mod download
make dev    # Hot-reload with Air
# Or:
make run    # Plain go run
```

Server starts at `http://localhost:8080`.

### 5. Verify

```bash
curl http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@test.com","password":"123456"}'
```

Swagger UI: http://localhost:8080/swagger/index.html

## Makefile Commands

```bash
make build    # Build binary to bin/server
make run      # go run .
make dev      # Air hot-reload (watches .go files)
make test     # go test -race ./...
make lint     # golangci-lint run
make clean    # Remove bin/
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `quizuser` | Database user |
| `DB_PASSWORD` | `quizpass` | Database password |
| `DB_NAME` | `quizdb` | Database name |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `JWT_SECRET` | `default-secret-change-me` | JWT signing key |
| `PORT` | `8080` | Server port |
| `CORS_ORIGINS` | `http://localhost:5175,http://localhost:3000` | Allowed CORS origins |

## Seed Data

Auto-seeded on first run:

| Email | Password | Role |
|-------|----------|------|
| `admin@test.com` | `123456` | admin |
| `player1@test.com` | `123456` | player |
| `player2@test.com` | `123456` | player |

Plus 2 sample quizzes: "General Knowledge (Live)" code `GENKNOW`, "Science & Tech (Self-paced)" code `SCITECH`.

---

## Troubleshooting

### `go: go.mod requires go >= 1.25`

**Cause:** Your Go version is too old.

```bash
go version
# If < 1.25:
# macOS
brew install go
# Or download from https://go.dev/dl/
```

### `failed to connect to database: ...`

**Cause:** PostgreSQL is not running or credentials are wrong.

```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432
# Output should be: "localhost:5432 - accepting connections"

# If not running:
# macOS
brew services start postgresql@16
# Linux
sudo systemctl start postgresql

# If using docker-compose (port 5433 on host):
docker compose up postgres -d
# Set in .env:
DB_PORT=5433
```

**If database doesn't exist:**

```bash
psql -U postgres -c "CREATE DATABASE quizdb;"
# Or with specific user:
psql -U postgres -c "CREATE USER quizuser WITH PASSWORD 'quizpass';"
psql -U postgres -c "CREATE DATABASE quizdb OWNER quizuser;"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE quizdb TO quizuser;"
```

**If authentication fails (`password authentication failed`):**

```bash
# Check pg_hba.conf allows your user. On macOS:
cat /usr/local/var/postgresql@16/pg_hba.conf
# Change "md5" to "trust" for local development, then restart:
brew services restart postgresql@16
```

### `failed to connect to redis: ...`

**Cause:** Redis is not running.

```bash
# Check if Redis is running
redis-cli ping
# Output should be: "PONG"

# If not running:
# macOS
brew services start redis
# Linux
sudo systemctl start redis

# If using docker-compose (port 6380 on host):
docker compose up redis -d
# Set in .env:
REDIS_ADDR=localhost:6380
```

### `address already in use :8080`

**Cause:** Another process is using port 8080.

```bash
# Find the process
lsof -i :8080
# Kill it
kill -9 <PID>
# Or change port in .env:
PORT=8081
```

### `go: no required module provides package ...`

**Cause:** Dependencies not downloaded or go.sum mismatch.

```bash
go mod download
go mod tidy
```

### `air: command not found`

**Cause:** Air is not installed.

```bash
go install github.com/air-verse/air@latest
# Make sure $GOPATH/bin is in your PATH:
export PATH=$PATH:$(go env GOPATH)/bin
```

### WebSocket connection fails (`ws://localhost:8080/ws/CODE`)

**Possible causes:**

1. **Quiz not started** — Quiz must be in `active` status. Call `POST /api/quizzes/:id/start` first.

2. **Invalid token** — For authenticated users, pass `?token=JWT` query param. Token expires in 15 minutes.

3. **Username taken** — For guests (`?guest=Name`), each username must be unique per quiz session.

4. **CORS blocking** — Ensure frontend origin is in `CORS_ORIGINS` env variable.

```bash
# Test WebSocket with wscat:
npm install -g wscat

# As guest:
wscat -c "ws://localhost:8080/ws/QUIZCODE?guest=TestPlayer"

# As authenticated user:
TOKEN=$(curl -s http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@test.com","password":"123456"}' | jq -r '.data.token')
wscat -c "ws://localhost:8080/ws/QUIZCODE?token=$TOKEN"
```

### `golangci-lint` errors

```bash
# Install golangci-lint v2
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Or via brew
brew install golangci-lint

# Run
golangci-lint run ./...
```

### Docker Compose — full stack

```bash
# From project root
docker compose up -d

# Check logs
docker compose logs backend -f

# Rebuild after code changes
docker compose up --build backend -d
```

> When using docker-compose, services communicate internally: `DB_HOST=postgres`, `REDIS_ADDR=redis:6379`. Host ports are `5433` (postgres), `6380` (redis), `8080` (backend), `5175` (frontend).

---

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for design patterns, module structure, DB schema, WebSocket protocol, and API reference.
