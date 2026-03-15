# Quiz Platform - Frontend

React-based frontend for the real-time quiz platform.

## Tech Stack

- React 19 + TypeScript
- React Router v7
- Axios (HTTP client)
- WebSocket (native) for real-time quiz play
- Nginx (production serving)

## Prerequisites

- Node.js 20+
- Backend running at `http://localhost:8080`

## Getting Started

```bash
# Install dependencies
npm install

# Start dev server (port 3000)
npm start
```

The app connects to the backend API at `http://localhost:8080/api` by default.
Override with the `REACT_APP_API_URL` environment variable.

## Docker

```bash
# Build and run via docker compose (from project root)
docker compose up --build

# Frontend will be available at http://localhost:5175
```

## Default Test Accounts

After backend starts, these seed accounts are available:

| Role   | Email              | Password |
|--------|--------------------|----------|
| Admin  | admin@test.com     | 123456   |
| Player | player1@test.com   | 123456   |
| Player | player2@test.com   | 123456   |

## Project Structure

```
src/
├── api/
│   └── http.ts           # Axios instance + API functions
├── components/
│   ├── Navbar.tsx         # Top navigation bar
│   ├── QuestionCard.tsx   # Quiz question with option buttons
│   ├── Leaderboard.tsx    # Live leaderboard sidebar
│   └── Timer.tsx          # Countdown timer bar
├── context/
│   └── AuthContext.tsx     # Auth state (login/register/logout)
├── hooks/
│   └── useWebSocket.ts    # WebSocket hook for quiz play
├── pages/
│   ├── LoginPage.tsx      # Login / Register
│   ├── DashboardPage.tsx  # Quiz management (admin)
│   ├── QuestionsPage.tsx  # Add/remove questions for a quiz
│   ├── JoinPage.tsx       # Enter quiz code to join
│   └── QuizPlayPage.tsx   # Live quiz play + leaderboard
├── App.tsx                # Router setup
└── index.tsx              # Entry point
```

## Usage Flow

### Admin
1. Login with admin account
2. **Dashboard** -> Create a quiz (set title + time per question)
3. Click **Questions** -> Add questions with options, mark correct answer
4. Click **Start** to make the quiz live
5. Share the quiz code with players
6. Join the quiz yourself to control question flow (Next Question)
7. Click **Finish** when done

### Player
1. Login or register as player
2. Go to **Join Quiz** -> Enter quiz code
3. Wait in lobby until host starts
4. Answer questions before time runs out
5. See live leaderboard and final results

## API Endpoints Used

| Method | Endpoint                       | Description          |
|--------|--------------------------------|----------------------|
| POST   | /api/auth/register             | Register user        |
| POST   | /api/auth/login                | Login                |
| GET    | /api/quizzes                   | List quizzes         |
| POST   | /api/quizzes                   | Create quiz          |
| DELETE | /api/quizzes/:id               | Delete quiz          |
| POST   | /api/quizzes/:id/start         | Start quiz           |
| POST   | /api/quizzes/:id/finish        | Finish quiz          |
| GET    | /api/quizzes/join/:code        | Join by code         |
| GET    | /api/quizzes/:id/questions     | List questions       |
| POST   | /api/quizzes/:id/questions     | Create question      |
| PUT    | /api/questions/:qid            | Update question      |
| DELETE | /api/questions/:qid            | Delete question      |
| WS     | /ws/:code                      | WebSocket quiz play  |
