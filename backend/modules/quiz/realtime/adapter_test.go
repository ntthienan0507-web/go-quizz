package realtime

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/chungnguyen/quizz-backend/modules/quiz/question"
	"github.com/chungnguyen/quizz-backend/pkg/ws"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alicebob/miniredis/v2"
)

func newTestAdapters(t *testing.T) (*ScorerAdapter, *LeaderboardAdapter, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	redisSvc := NewRedisService(client)
	scoringSvc := NewScoringService(redisSvc)
	return NewScorerAdapter(scoringSvc), NewLeaderboardAdapter(redisSvc), mr
}

func TestScorerAdapter_ProcessAnswer(t *testing.T) {
	scorer, _, _ := newTestAdapters(t)
	ctx := context.Background()

	qData := ws.QuestionData{
		Text:       "Test?",
		Options:    json.RawMessage(`["A","B","C"]`),
		CorrectIdx: 1,
		Points:     10,
	}

	result, err := scorer.ProcessAnswer(ctx, "QUIZ1", "u1", 0, 1, qData, 30, 30)
	require.NoError(t, err)
	assert.True(t, result.IsCorrect)
	assert.Equal(t, 10, result.PointsAwarded)
}

func TestLeaderboardAdapter_GetLeaderboard(t *testing.T) {
	scorer, lb, _ := newTestAdapters(t)
	ctx := context.Background()

	// Score some answers
	qData := ws.QuestionData{
		Text:       "Q?",
		Options:    json.RawMessage(`["A","B"]`),
		CorrectIdx: 0,
		Points:     10,
	}
	_, _ = scorer.ProcessAnswer(ctx, "QUIZ1", "u1", 0, 0, qData, 30, 30)
	_, _ = scorer.ProcessAnswer(ctx, "QUIZ1", "u2", 0, 0, qData, 20, 30)

	_ = lb.SetUsername(ctx, "QUIZ1", "u1", "Alice")
	_ = lb.SetUsername(ctx, "QUIZ1", "u2", "Bob")

	entries, err := lb.GetLeaderboard(ctx, "QUIZ1", 10)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "Alice", entries[0].Username)
}

func TestToQuestionDataList(t *testing.T) {
	questions := []question.Question{
		{
			ID:         uuid.New(),
			Text:       "Q1",
			Options:    json.RawMessage(`["A","B"]`),
			CorrectIdx: 0,
			Points:     10,
		},
		{
			ID:         uuid.New(),
			Text:       "Q2",
			Options:    json.RawMessage(`["C","D"]`),
			CorrectIdx: 1,
			Points:     20,
		},
	}

	result := ToQuestionDataList(questions)

	assert.Len(t, result, 2)
	assert.Equal(t, "Q1", result[0].Text)
	assert.Equal(t, 10, result[0].Points)
	assert.Equal(t, "Q2", result[1].Text)
	assert.Equal(t, 1, result[1].CorrectIdx)
}
