package realtime

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedis(t *testing.T) (*RedisService, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewRedisService(client), mr
}

func TestUpdateScoreAndGetLeaderboard(t *testing.T) {
	svc, _ := newTestRedis(t)
	ctx := context.Background()

	require.NoError(t, svc.UpdateScore(ctx, "ABC", "user1", 10))
	require.NoError(t, svc.UpdateScore(ctx, "ABC", "user2", 20))
	require.NoError(t, svc.UpdateScore(ctx, "ABC", "user1", 5))

	entries, err := svc.GetLeaderboard(ctx, "ABC", 10)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "user2", entries[0].UserID)
	assert.Equal(t, float64(20), entries[0].Score)
	assert.Equal(t, 1, entries[0].Rank)
	assert.Equal(t, "user1", entries[1].UserID)
	assert.Equal(t, float64(15), entries[1].Score)
	assert.Equal(t, 2, entries[1].Rank)
}

func TestSetUsernameAppearsInLeaderboard(t *testing.T) {
	svc, _ := newTestRedis(t)
	ctx := context.Background()

	require.NoError(t, svc.UpdateScore(ctx, "ABC", "u1", 10))
	require.NoError(t, svc.SetUsername(ctx, "ABC", "u1", "Alice"))

	entries, err := svc.GetLeaderboard(ctx, "ABC", 10)
	require.NoError(t, err)
	assert.Equal(t, "Alice", entries[0].Username)
}

func TestGetUserScore(t *testing.T) {
	svc, _ := newTestRedis(t)
	ctx := context.Background()

	require.NoError(t, svc.UpdateScore(ctx, "ABC", "u1", 42))

	score, err := svc.GetUserScore(ctx, "ABC", "u1")
	require.NoError(t, err)
	assert.Equal(t, float64(42), score)
}

func TestTrackAnswerAtomic(t *testing.T) {
	svc, _ := newTestRedis(t)
	ctx := context.Background()

	// First call should return true (newly added)
	added, err := svc.TrackAnswerAtomic(ctx, "ABC", 0, "u1")
	require.NoError(t, err)
	assert.True(t, added)

	// Second call should return false (already answered)
	added, err = svc.TrackAnswerAtomic(ctx, "ABC", 0, "u1")
	require.NoError(t, err)
	assert.False(t, added)

	// Different question index should return true
	added, err = svc.TrackAnswerAtomic(ctx, "ABC", 1, "u1")
	require.NoError(t, err)
	assert.True(t, added)
}

func TestGetAllScores(t *testing.T) {
	svc, _ := newTestRedis(t)
	ctx := context.Background()

	require.NoError(t, svc.UpdateScore(ctx, "ABC", "u1", 10))
	require.NoError(t, svc.UpdateScore(ctx, "ABC", "u2", 20))
	require.NoError(t, svc.UpdateScore(ctx, "ABC", "u3", 30))

	entries, err := svc.GetAllScores(ctx, "ABC")
	require.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.Equal(t, "u3", entries[0].UserID)
}

func TestCleanup(t *testing.T) {
	svc, mr := newTestRedis(t)
	ctx := context.Background()

	require.NoError(t, svc.UpdateScore(ctx, "ABC", "u1", 10))
	require.NoError(t, svc.SetUsername(ctx, "ABC", "u1", "Alice"))
	_, trackErr := svc.TrackAnswerAtomic(ctx, "ABC", 0, "u1")
	require.NoError(t, trackErr)

	keys := mr.Keys()
	assert.NotEmpty(t, keys)

	require.NoError(t, svc.Cleanup(ctx, "ABC"))

	keys = mr.Keys()
	assert.Empty(t, keys)
}
