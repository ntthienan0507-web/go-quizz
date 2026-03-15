package realtime

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/chungnguyen/quizz-backend/modules/quiz/question"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestScoringService(t *testing.T) (*ScoringService, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	redisSvc := NewRedisService(client)
	return NewScoringService(redisSvc), mr
}

func makeQuestion(correctIdx, points int) question.Question {
	opts, _ := json.Marshal([]string{"A", "B", "C", "D"})
	return question.Question{
		Text:       "What is 1+1?",
		Options:    opts,
		CorrectIdx: correctIdx,
		Points:     points,
	}
}

func TestProcessAnswer_Correct(t *testing.T) {
	svc, _ := newTestScoringService(t)
	ctx := context.Background()

	q := makeQuestion(1, 10)
	result, err := svc.ProcessAnswer(ctx, "ABC", "u1", 0, 1, q, 30, 30)

	require.NoError(t, err)
	assert.True(t, result.IsCorrect)
	assert.Equal(t, 10, result.PointsAwarded) // full time = full points
	assert.Equal(t, 1, result.CorrectIdx)
	assert.Equal(t, float64(10), result.YourTotal)
}

func TestProcessAnswer_Wrong(t *testing.T) {
	svc, _ := newTestScoringService(t)
	ctx := context.Background()

	q := makeQuestion(1, 10)
	result, err := svc.ProcessAnswer(ctx, "ABC", "u1", 0, 2, q, 30, 30)

	require.NoError(t, err)
	assert.False(t, result.IsCorrect)
	assert.Equal(t, 0, result.PointsAwarded)
	assert.Equal(t, float64(0), result.YourTotal)
}

func TestProcessAnswer_PartialTime(t *testing.T) {
	svc, _ := newTestScoringService(t)
	ctx := context.Background()

	q := makeQuestion(0, 10)
	// Half time remaining = ~5 points
	result, err := svc.ProcessAnswer(ctx, "ABC", "u1", 0, 0, q, 15, 30)

	require.NoError(t, err)
	assert.True(t, result.IsCorrect)
	assert.Equal(t, 5, result.PointsAwarded)
}

func TestProcessAnswer_MinimumOnePoint(t *testing.T) {
	svc, _ := newTestScoringService(t)
	ctx := context.Background()

	q := makeQuestion(0, 10)
	// Almost no time left
	result, err := svc.ProcessAnswer(ctx, "ABC", "u1", 0, 0, q, 1, 300)

	require.NoError(t, err)
	assert.True(t, result.IsCorrect)
	assert.Equal(t, 1, result.PointsAwarded) // minimum 1 point
}

func TestProcessAnswer_DuplicateAnswer(t *testing.T) {
	svc, _ := newTestScoringService(t)
	ctx := context.Background()

	q := makeQuestion(1, 10)

	_, err := svc.ProcessAnswer(ctx, "ABC", "u1", 0, 1, q, 30, 30)
	require.NoError(t, err)

	// Second answer to same question
	_, err = svc.ProcessAnswer(ctx, "ABC", "u1", 0, 1, q, 30, 30)
	assert.ErrorIs(t, err, ErrAlreadyAnswered)
}

func TestProcessAnswer_DifferentQuestionsAllowed(t *testing.T) {
	svc, _ := newTestScoringService(t)
	ctx := context.Background()

	q := makeQuestion(0, 10)

	_, err := svc.ProcessAnswer(ctx, "ABC", "u1", 0, 0, q, 30, 30)
	require.NoError(t, err)

	// Different question index should work
	result, err := svc.ProcessAnswer(ctx, "ABC", "u1", 1, 0, q, 20, 30)
	require.NoError(t, err)
	assert.True(t, result.IsCorrect)
	assert.Equal(t, float64(17), result.YourTotal) // 10 + 7
}

func TestProcessAnswer_AccumulatesScore(t *testing.T) {
	svc, _ := newTestScoringService(t)
	ctx := context.Background()

	q := makeQuestion(0, 10)

	r1, _ := svc.ProcessAnswer(ctx, "ABC", "u1", 0, 0, q, 30, 30)
	assert.Equal(t, float64(10), r1.YourTotal)

	r2, _ := svc.ProcessAnswer(ctx, "ABC", "u1", 1, 0, q, 30, 30)
	assert.Equal(t, float64(20), r2.YourTotal)

	r3, _ := svc.ProcessAnswer(ctx, "ABC", "u1", 2, 0, q, 15, 30)
	assert.Equal(t, float64(25), r3.YourTotal) // 10 + 10 + 5
}
