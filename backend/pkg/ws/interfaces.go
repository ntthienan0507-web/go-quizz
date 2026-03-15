package ws

import (
	"context"
	"encoding/json"
)

// QuestionData represents a question for the hub (interface-free).
type QuestionData struct {
	Text       string
	Options    json.RawMessage
	CorrectIdx int
	Points     int
}

// ScoringResult is the result of processing an answer.
type ScoringResult struct {
	IsCorrect     bool
	PointsAwarded int
	CorrectIdx    int
	YourTotal     float64
}

// Scorer validates answers and updates scores.
type Scorer interface {
	ProcessAnswer(ctx context.Context, quizCode, userID string, questionIdx, selectedIdx int, question QuestionData, timeRemaining, timeTotal int) (*ScoringResult, error)
}

// LeaderboardProvider fetches leaderboard data.
type LeaderboardProvider interface {
	GetLeaderboard(ctx context.Context, quizCode string, top int) ([]LeaderboardEntryData, error)
	SetUsername(ctx context.Context, quizCode, userID, username string) error
}

// LeaderboardEntryData is a leaderboard row.
type LeaderboardEntryData struct {
	UserID   string
	Username string
	Score    float64
	Rank     int
}
