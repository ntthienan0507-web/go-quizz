# Backend Architecture & Design Patterns

All patterns are already implemented. Read before coding — don't reinvent solutions for problems that have already been solved.

---

## 1. Three-Layer Architecture (Controller → Service → Repository)

**Problem:** Business logic in controllers — coupled with HTTP, hard to test, hard to reuse.

**Solution:** Every module follows 3 layers. Each layer has exactly one job.

```
HTTP Request
     │
Controller    Parse request, validate input, format response
     │
  Service     Business logic, orchestration, validation rules
     │
Repository    Data access (PostgreSQL via GORM, Redis)
     │
  Database
```

```go
// Controller — only HTTP concerns
func (ctrl *Controller) Create(c *gin.Context) {
    var req CreateQuizRequest
    if err := c.ShouldBindJSON(&req); err != nil { ... }
    userID, _ := uuid.Parse(c.GetString("userID"))

    quiz, err := ctrl.service.Create(ctx, userID, req.Title, req.Mode, req.TimePerQuestion)
    response.Created(c, ToQuizResponse(quiz))
}

// Service — business logic
func (s *Service) Create(ctx context.Context, userID uuid.UUID, title, mode string, tpq int) (*Quiz, error) {
    if tpq == 0 { tpq = 30 }         // default
    if mode != ModeSelfPaced { mode = ModeLive }  // validation
    quiz := &Quiz{QuizCode: generateCode(), ...}  // domain logic
    return quiz, s.repo.CreateQuiz(ctx, quiz)
}

// Repository — data access only
func (r *GormRepository) CreateQuiz(ctx context.Context, quiz *Quiz) error {
    return r.db.WithContext(ctx).Create(quiz).Error
}
```

**Rules:**
- Controllers NEVER call repositories directly
- Services hold ALL business logic (defaults, validation, orchestration)
- Repositories are interfaces — swap implementations for testing

**When to use:** Every module, no exceptions.
**When NOT to use:** Never skip — even simple CRUD benefits from testability.

**Files:** `modules/*/controller.go`, `modules/*/service.go`, `modules/*/repository.go`

---

## 2. Repository Interface Pattern

**Problem:** Service is coupled to GORM. Can't unit test without a database. Can't switch ORM.

**Solution:** Interface in `repository.go`, GORM implementation below. Service depends on the interface.

```go
// Interface — service depends on this
type Repository interface {
    CreateQuiz(ctx context.Context, quiz *Quiz) error
    GetQuizByID(ctx context.Context, id uuid.UUID) (*Quiz, error)
    GetQuizByCode(ctx context.Context, code string) (*Quiz, error)
    ListQuizzesByUser(ctx context.Context, userID uuid.UUID) ([]Quiz, error)
    UpdateQuiz(ctx context.Context, quiz *Quiz) error
    DeleteQuiz(ctx context.Context, id uuid.UUID) error
    BatchCreateResults(ctx context.Context, results []QuizResult) error
}

// Implementation — hidden behind interface
type GormRepository struct { db *gorm.DB }
func NewRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }
```

All modules follow this:

| Module | Interface | Implementation |
|--------|-----------|----------------|
| auth | `Repository` | `repository` (unexported) |
| quiz | `Repository` | `GormRepository` |
| question | `Repository` | `GormRepository` |
| player | `IRepository` | `Repository` |

**When to use:** All database access.
**When NOT to use:** In-memory data, config lookup.

**Files:** `modules/*/repository.go`

---

## 3. Dependency Injection

**Problem:** Global variables, hidden dependencies, untestable code.

**Solution:** Constructor injection. All dependencies passed through `New*()`. Single wiring point in `main.go`.

```go
// Wiring in main.go
authRepo       → authService       → authController       → authRoutes

questionRepo   → questionService   → questionController   → questionRoutes

quizRepo       ─┐
questionRepo   ─┤
redisService   ─┼→ quizService     → quizController       → quizRoutes
scoringService ─┤
hubManager     ─┘

playerRepo     ─┐
authService    ─┴→ playerService   → playerController     → playerRoutes
```

```go
// Each constructor declares exactly what it needs
func NewService(
    repo Repository,
    questionRepo question.Repository,
    redisService *realtime.RedisService,
    scoringService *realtime.ScoringService,
    hubManager *ws.HubManager,
) *Service
```

**When to use:** All structs that have dependencies.
**When NOT to use:** Pure functions that don't need state.

**File:** `main.go` — single wiring point

---

## 4. Adapter Pattern

**Problem:** `ws.Hub` needs a `Scorer` interface. `ScoringService` has the logic but different method signatures. Can't import `ws` from `realtime` (circular dependency).

**Solution:** Adapter bridges the gap — implements the interface, delegates to the concrete service.

```go
// ws package defines the interface
type Scorer interface {
    ProcessAnswer(ctx context.Context, quizCode, userID string, ...) (*ScoringResult, error)
}

// realtime package has the implementation with different types
type ScoringService struct { ... }
func (s *ScoringService) ProcessAnswer(..., q question.Question, ...) (*ScoringResult, error)

// Adapter bridges them
type ScorerAdapter struct { svc *ScoringService }

func (a *ScorerAdapter) ProcessAnswer(ctx context.Context, ..., q ws.QuestionData, ...) (*ws.ScoringResult, error) {
    qModel := question.Question{Text: q.Text, Options: q.Options, ...}
    result, err := a.svc.ProcessAnswer(ctx, ..., qModel, ...)
    return &ws.ScoringResult{IsCorrect: result.IsCorrect, ...}, err
}
```

Two adapters exist:

| Adapter | Bridges | Interface |
|---------|---------|-----------|
| `ScorerAdapter` | `ScoringService` → `ws.Scorer` | Answer processing |
| `LeaderboardAdapter` | `RedisService` → `ws.LeaderboardProvider` | Leaderboard queries |

**When to use:** Module A needs module B's logic but they can't import each other, or have incompatible types.
**When NOT to use:** Same package — just call directly.

**File:** `modules/quiz/realtime/adapter.go`

---

## 5. Hub Pattern (Pub/Sub per Room)

**Problem:** Real-time quiz needs to broadcast questions and collect answers from N players simultaneously. Each quiz is an isolated room.

**Solution:** One Hub per active quiz. Hub manages client registration, message broadcasting, and game state.

```
Player A ──┐                    ┌── Player A
Player B ──┼── WebSocket ── Hub ┼── Player B    (broadcast)
Player C ──┘     (per quiz)     └── Player C

Host ── "next_question" ── Hub ── broadcast "new_question" to all
Player ── "submit_answer" ── Hub ── score ── send "answer_result" to player
                                          ── broadcast "leaderboard_update" to all
```

```go
type Hub struct {
    clients       map[*Client]bool   // connected players
    broadcast     chan []byte         // fan-out channel
    register      chan *Client        // join
    unregister    chan *Client        // leave
    questions     []QuestionData     // quiz content
    currentQ      int                // current question index (live mode)
    answeredUsers map[string]bool    // who answered current question
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:    // player joins
        case client := <-h.unregister:  // player leaves
        case message := <-h.broadcast:  // send to all clients
        }
    }
}
```

**Two modes:**
- `live` — host controls question flow via `next_question`, all players see same question
- `self_paced` — each player progresses independently via per-client `currentQ`

**When to use:** Real-time multiplayer features with room isolation.
**When NOT to use:** Simple request-response (use REST).

**Files:** `pkg/ws/hub.go`, `pkg/ws/client.go`, `pkg/ws/handler.go`

---

## 6. Dual Auth Strategy (Token + Guest)

**Problem:** Registered users authenticate with JWT. But guests should play without an account.

**Solution:** WebSocket handler accepts two auth modes via query params.

```go
// handler.go
func (h *WSHandler) Handle(c *gin.Context) {
    token := c.Query("token")
    guestName := c.Query("guest")

    if token != "" {
        // Validate JWT → extract userID, username, role
        claims, err := h.authService.ValidateToken(token)
        client = &Client{UserID: claims.UserID, Username: claims.Username, IsGuest: false}
    } else if guestName != "" {
        // Guest — generate temp UUID, use provided name
        client = &Client{UserID: uuid.New().String(), Username: guestName, IsGuest: true}
    } else {
        // Reject — no auth
    }

    // Check username uniqueness within the hub
    if hub.IsUsernameTaken(client.Username) { ... }
}
```

**Frontend usage:**
```typescript
// Authenticated user
const url = `ws://localhost:8080/ws/${code}?token=${jwt}`;

// Guest
const url = `ws://localhost:8080/ws/${code}?guest=${encodeURIComponent(name)}`;
```

**When to use:** Features that need both authenticated and anonymous access.
**When NOT to use:** Admin-only features (use token-only auth).

**File:** `pkg/ws/handler.go`

---

## 7. JWT + Refresh Token Rotation

**Problem:** Long-lived tokens are risky if stolen. Short-lived tokens force frequent re-login.

**Solution:** Short access token (15 min) + long refresh token (7 days). Refresh token is single-use (rotation).

```
Login → access_token (15 min) + refresh_token (7 days, stored in DB)

access_token expired → POST /auth/refresh { refresh_token }
  → delete old refresh_token
  → issue new access_token + new refresh_token
  → return both
```

```go
func (s *Service) RefreshTokens(ctx context.Context, refreshTokenStr string) (string, string, *User, error) {
    rt, _ := s.repo.GetRefreshToken(ctx, refreshTokenStr)
    if time.Now().After(rt.ExpiresAt) { return error }

    s.repo.DeleteRefreshToken(ctx, refreshTokenStr)  // rotation — old token is dead

    accessToken, _ := s.generateToken(user)           // new 15-min token
    newRefreshToken, _ := s.generateRefreshToken(ctx, user.ID)  // new 7-day token

    return accessToken, newRefreshToken, user, nil
}
```

**Security:**
- Stolen refresh token can only be used once — if attacker uses it, legitimate user's next refresh fails → detected
- Refresh tokens are stored in DB → can be revoked server-side
- Access tokens are stateless JWT → no DB lookup on every request

**When to use:** User authentication.
**When NOT to use:** Service-to-service auth (use API keys or mTLS).

**Files:** `modules/auth/service.go`, `modules/auth/repository.go`

---

## 8. Rate Limiting (Redis Sliding Window)

**Problem:** Brute-force login, API abuse, WebSocket connection flooding.

**Solution:** Redis `INCR` + `EXPIRE` per IP (or per user). Fail-open if Redis is down.

```go
func RateLimit(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("rl:%s:%s", c.FullPath(), c.ClientIP())
        count, _ := rdb.Incr(ctx, key).Result()
        if count == 1 { rdb.Expire(ctx, key, cfg.Window) }

        if int(count) > cfg.Max {
            c.Header("Retry-After", ...)
            c.AbortWithStatusJSON(429, ...)
            return
        }
        c.Next()
    }
}
```

Applied per endpoint group:

| Scope | Limit | Window | Applied To |
|-------|-------|--------|------------|
| Auth | 10 req | 1 min | `/api/auth/*` (login, register, refresh) |
| API | 60 req | 1 min | All `/api/*` endpoints |
| WebSocket | 5 conn | 1 min | `/ws/:code` |

**Two variants:**
- `RateLimit` — keys on client IP (public endpoints)
- `RateLimitByUser` — keys on authenticated userID, falls back to IP

**Response headers:** `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After`

**When to use:** All public endpoints, especially auth.
**When NOT to use:** Internal service-to-service calls.

**File:** `pkg/middleware/ratelimit.go`

---

## 9. Middleware Chain

**Problem:** Cross-cutting concerns (auth, CORS, rate limiting) repeated in every handler.

**Solution:** Gin middleware stack — each middleware handles one concern.

```
Request → CORS → RateLimit (global) → Auth → RateLimit (per-route) → Handler
```

| Middleware | File | Scope |
|-----------|------|-------|
| CORS | `cors.go` | Global — cross-origin from `CORS_ORIGINS` env |
| RateLimit | `ratelimit.go` | Global (60/min) + per-group (auth: 10/min, ws: 5/min) |
| Auth | `auth.go` | Per-group — validate JWT, set userID/username/role in context |

```go
// Auth middleware extracts claims and sets context values
func Auth(authService *auth.Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, _ := authService.ValidateToken(bearerToken)
        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

**When to create new middleware:** Concern repeats across 3+ routes.
**When NOT to use:** One-off logic for a single handler (put it in the handler).

**File:** `pkg/middleware/`

---

## 10. Redis as Real-Time State Store

**Problem:** Scoring and leaderboards need sub-millisecond updates for N concurrent players. PostgreSQL too slow for per-answer writes during live quiz.

**Solution:** Redis sorted sets for leaderboard, hash maps for usernames, sets for answer tracking. Persist to PostgreSQL only when quiz finishes.

```
During quiz (Redis):
  ZADD quiz:{code}:scores {points} {userID}     ← leaderboard
  HSET quiz:{code}:usernames {userID} {name}     ← display names
  SADD quiz:{code}:q:{idx}:answered {userID}     ← prevent double-answer

Quiz finishes (Redis → PostgreSQL):
  ZRANGE quiz:{code}:scores → []QuizResult → BatchCreateResults()
  DEL quiz:{code}:*                         → cleanup
```

```go
// Time-based scoring
func (s *ScoringService) ProcessAnswer(...) (*ScoringResult, error) {
    // Check already answered
    if s.redis.HasAnswered(ctx, quizCode, questionIdx, userID) {
        return nil, ErrAlreadyAnswered
    }
    s.redis.MarkAnswered(ctx, quizCode, questionIdx, userID)

    // Calculate points: faster answer = more points
    if selectedIdx == question.CorrectIdx {
        points = max(basePoints * timeRemaining / timeTotal, 1)
    }

    s.redis.AddScore(ctx, quizCode, userID, points)
    return &ScoringResult{IsCorrect: true, PointsAwarded: points, ...}, nil
}
```

**Scoring formula:** `points = max(basePoints × timeRemaining / timeTotal, 1)`

**When to use:** Real-time features with high write throughput that can be persisted later.
**When NOT to use:** Data that must be immediately durable (use PostgreSQL directly).

**Files:** `modules/quiz/realtime/redis.go`, `modules/quiz/realtime/scoring.go`

---

## 11. Quiz Lifecycle State Machine

**Problem:** Quiz has states (draft → active → finished) with complex transitions. Each transition involves multiple systems (DB, Redis, WebSocket).

**Solution:** Service methods enforce transitions and orchestrate side effects.

```
draft ──Start()──► active ──Finish()──► finished
          │                    │
          ├─ Validate questions exist   ├─ Fetch Redis scores
          ├─ Update DB status           ├─ Transform → QuizResult models
          ├─ Create WS Hub              ├─ Batch persist to PostgreSQL
          └─ Start goroutine            ├─ Update DB status
                                        ├─ Cleanup Redis keys
                                        └─ Remove WS Hub
```

```go
func (s *Service) Start(ctx context.Context, id uuid.UUID) (*Quiz, error) {
    quiz, _ := s.repo.GetQuizByID(ctx, id)
    if quiz.Status != StatusDraft { return nil, fmt.Errorf("not in draft") }

    questions, _ := s.questionRepo.ListByQuizID(ctx, id)
    if len(questions) == 0 { return nil, fmt.Errorf("no questions") }

    quiz.Status = StatusActive
    s.repo.UpdateQuiz(ctx, quiz)

    hub := ws.NewHub(quiz.QuizCode, ..., s.scorerAdapter, s.leaderboardAdapter)
    s.hubManager.CreateHub(hub)

    return quiz, nil
}
```

**When to use:** Entities with lifecycle states and complex transitions.
**When NOT to use:** Simple CRUD without state transitions.

**File:** `modules/quiz/service.go`

---

## 12. Response Envelope

**Problem:** Inconsistent API responses make frontend parsing fragile.

**Solution:** All responses follow a standard envelope with typed helpers.

```go
// Success
response.OK(c, data)       // { "status": "success", "data": {...} }
response.Created(c, data)  // 201 + same envelope
response.OKList(c, items, total)  // { "status": "success", "data": { "items": [...], "total": N } }

// Errors
response.BadRequest(c, "msg")     // 400
response.Unauthorized(c, "msg")   // 401
response.NotFound(c, "msg")       // 404
response.Conflict(c, "msg")       // 409
response.InternalError(c, "msg")  // 500
```

**When to use:** All HTTP responses.
**When NOT to use:** WebSocket messages (use WS message types).

**File:** `pkg/response/response.go`

---

## Quick Reference — "I need..."

| Need | Use pattern | File |
|------|------------|------|
| Add new module | 3-Layer (#1) | `modules/*/` — models, types, repository, service, controller, routes |
| Database access | Repository Interface (#2) | `modules/*/repository.go` |
| Wire dependencies | DI (#3) | `main.go` |
| Bridge incompatible modules | Adapter (#4) | `modules/quiz/realtime/adapter.go` |
| Real-time multiplayer room | Hub (#5) | `pkg/ws/hub.go` |
| Auth + guest access | Dual Auth (#6) | `pkg/ws/handler.go` |
| User login/session | JWT + Refresh (#7) | `modules/auth/service.go` |
| Prevent brute-force/abuse | Rate Limit (#8) | `pkg/middleware/ratelimit.go` |
| Cross-cutting HTTP concern | Middleware (#9) | `pkg/middleware/` |
| Fast scoring/leaderboard | Redis State (#10) | `modules/quiz/realtime/` |
| Entity with lifecycle states | State Machine (#11) | `modules/quiz/service.go` (Start, Finish) |
| Consistent API responses | Response Envelope (#12) | `pkg/response/response.go` |

---

## Project Structure

```
backend/
├── main.go                           # DI wiring point
├── Makefile                          # build, run, test, lint, dev
├── .air.toml                         # Hot-reload config
├── .golangci.yml                     # Linter config (v2)
├── Dockerfile
│
├── pkg/                              # Shared infrastructure
│   ├── config/config.go              # Env-based configuration
│   ├── db/
│   │   ├── postgres.go               # GORM connection + AutoMigrate
│   │   ├── redis.go                  # Redis client
│   │   └── seed/                     # Initial data (per-table files)
│   ├── middleware/
│   │   ├── auth.go                   # JWT Bearer validation
│   │   ├── cors.go                   # CORS from env
│   │   └── ratelimit.go             # Redis-based rate limiting
│   ├── response/response.go         # Generic response helpers
│   └── ws/
│       ├── interfaces.go             # Scorer, LeaderboardProvider
│       ├── hub.go                    # Per-quiz hub (broadcast, game state)
│       ├── client.go                 # WS client (read/write pumps)
│       ├── message.go               # WS message types
│       └── handler.go               # HTTP → WS upgrade + dual auth
│
├── modules/
│   ├── auth/                         # controller → service → repository
│   │   ├── models.go                 # User, RefreshToken
│   │   ├── types.go                  # DTOs
│   │   ├── repository.go            # Interface + GORM impl
│   │   ├── service.go               # Auth logic (bcrypt, JWT, refresh rotation)
│   │   ├── controller.go            # /register, /login, /refresh
│   │   └── routes.go
│   │
│   ├── quiz/                         # controller → service → repository
│   │   ├── models.go                 # Quiz, QuizResult, UserAnswer
│   │   ├── types.go                  # DTOs + transformers
│   │   ├── repository.go            # Interface + GORM impl
│   │   ├── service.go               # Create, Start, Finish (state machine)
│   │   ├── controller.go
│   │   ├── routes.go
│   │   ├── question/                 # Sub-module: controller → service → repository
│   │   │   ├── models.go            # Question (JSONB options)
│   │   │   ├── types.go, repository.go, service.go, controller.go, routes.go
│   │   └── realtime/                 # No controller — internal services
│   │       ├── redis.go              # Sorted sets, usernames, answer tracking
│   │       ├── scoring.go           # Time-based scoring, double-answer guard
│   │       └── adapter.go           # Bridges to ws interfaces
│   │
│   └── player/                       # controller → service → repository
│       ├── types.go, repository.go, service.go, controller.go, routes.go
```

---

## Database Schema

```
┌──────────────┐     ┌──────────────────┐     ┌──────────────┐
│    users     │     │  refresh_tokens  │     │   quizzes    │
├──────────────┤     ├──────────────────┤     ├──────────────┤
│ id (UUID PK) │◄────│ user_id (FK)     │     │ id (UUID PK) │
│ username (UQ)│     │ token            │     │ title        │
│ email (UQ)   │     │ expires_at       │     │ quiz_code(UQ)│
│ password_hash│     └──────────────────┘     │ created_by   │──► users.id
│ role         │                              │ status       │ draft→active→finished
│ created_at   │     ┌──────────────────┐     │ mode         │ live | self_paced
│ updated_at   │     │   questions      │     │ time_per_q   │
└──────────────┘     ├──────────────────┤     └──────────────┘
                     │ id (UUID PK)     │           │
                     │ quiz_id (FK)     │───────────┘
                     │ text             │
                     │ options (JSONB)  │     ┌──────────────────┐
                     │ correct_idx      │     │  quiz_results    │
                     │ points           │     ├──────────────────┤
                     │ order_num        │     │ id (UUID PK)     │
                     └──────────────────┘     │ quiz_id (FK)     │──► quizzes.id
                                              │ user_id (FK)     │──► users.id
                     ┌──────────────────┐     │ score, rank      │
                     │  user_answers    │     │ finished_at      │
                     ├──────────────────┤     └──────────────────┘
                     │ id (UUID PK)     │
                     │ quiz_id (FK)     │──► quizzes.id
                     │ question_id (FK) │──► questions.id
                     │ user_id (FK)     │──► users.id
                     │ selected_idx     │
                     │ is_correct       │
                     │ answered_at      │
                     └──────────────────┘
```

---

## WebSocket Protocol

Connection: `GET /ws/:code?token=JWT` or `GET /ws/:code?guest=Name`

### Server → Client

| Type | Payload | When |
|------|---------|------|
| `welcome` | quiz_title, total_questions, participants, mode | On connect |
| `player_joined` | username, participant_count | Player joins |
| `player_left` | username, participant_count | Player disconnects |
| `new_question` | question_idx, text, options, time_limit | Host advances / self-paced |
| `answer_result` | is_correct, points_awarded, correct_idx, your_total | After answer |
| `answer_progress` | question_idx, answered_count, total_players | Any player answers (live) |
| `leaderboard_update` | rankings[] | After scoring changes |
| `quiz_finished` | rankings[] (final) | Quiz ends |
| `error` | message | Validation errors |

### Client → Server

| Type | Payload | Description |
|------|---------|-------------|
| `submit_answer` | question_idx, selected_idx, time_remaining | Player answers |
| `next_question` | — | Host advances (live) / player advances (self-paced) |
| `ping` | — | Keep-alive |

---

## API Endpoints

### Auth (Public, rate limited: 10/min)
```
POST /api/auth/register    { username, email, password, role }
POST /api/auth/login       { email, password } → { token, refresh_token, user }
POST /api/auth/refresh     { refresh_token }   → { token, refresh_token, user }
```

### Quizzes (Protected)
```
GET    /api/quizzes              List user's quizzes
POST   /api/quizzes              Create { title, mode, time_per_question }
GET    /api/quizzes/:id          Get quiz details
PUT    /api/quizzes/:id          Update { title?, time_per_question? }
DELETE /api/quizzes/:id          Delete quiz
POST   /api/quizzes/:id/start   Start → creates WS hub
POST   /api/quizzes/:id/finish  Finish → Redis scores → PostgreSQL
GET    /api/quizzes/join/:code   (Public) Validate quiz code
```

### Questions (Protected)
```
GET    /api/quizzes/:id/questions     List
POST   /api/quizzes/:id/questions     Create { text, options, correct_idx, points, order_num }
PUT    /api/questions/:qid            Update
DELETE /api/questions/:qid            Delete
```

### Player (Protected)
```
GET    /api/player/dashboard     Stats + global rank + recent
GET    /api/player/history       Paginated (?page=1&limit=10)
GET    /api/player/leaderboard   Global (?page=1&limit=20)
GET    /api/player/profile       User info + stats
PUT    /api/player/profile       Update { username?, email? }
```

### WebSocket (Rate limited: 5/min)
```
GET    /ws/:code?token=JWT       Authenticated
GET    /ws/:code?guest=Name      Guest
```
