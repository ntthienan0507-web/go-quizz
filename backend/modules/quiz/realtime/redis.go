package realtime

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const keyTTL = 24 * time.Hour

type RedisService struct {
	client *redis.Client
}

func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{client: client}
}

type LeaderboardEntry struct {
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Score    float64 `json:"score"`
	Rank     int     `json:"rank"`
}

func (rs *RedisService) leaderboardKey(quizCode string) string {
	return fmt.Sprintf("quiz:%s:leaderboard", quizCode)
}

func (rs *RedisService) usernameKey(quizCode string) string {
	return fmt.Sprintf("quiz:%s:usernames", quizCode)
}

func (rs *RedisService) answeredKey(quizCode string, questionIdx int) string {
	return fmt.Sprintf("quiz:%s:q:%d:answered", quizCode, questionIdx)
}

func (rs *RedisService) UpdateScore(ctx context.Context, quizCode, userID string, points float64) error {
	key := rs.leaderboardKey(quizCode)
	if err := rs.client.ZIncrBy(ctx, key, points, userID).Err(); err != nil {
		return err
	}
	rs.client.Expire(ctx, key, keyTTL)
	return nil
}

func (rs *RedisService) SetUsername(ctx context.Context, quizCode, userID, username string) error {
	key := rs.usernameKey(quizCode)
	if err := rs.client.HSet(ctx, key, userID, username).Err(); err != nil {
		return err
	}
	rs.client.Expire(ctx, key, keyTTL)
	return nil
}

func (rs *RedisService) GetLeaderboard(ctx context.Context, quizCode string, top int) ([]LeaderboardEntry, error) {
	stop := int64(top - 1)
	if top <= 0 {
		stop = -1 // all elements
	}
	results, err := rs.client.ZRevRangeWithScores(ctx, rs.leaderboardKey(quizCode), 0, stop).Result()
	if err != nil {
		return nil, err
	}

	usernames, _ := rs.client.HGetAll(ctx, rs.usernameKey(quizCode)).Result()

	entries := make([]LeaderboardEntry, len(results))
	for i, z := range results {
		userID := z.Member.(string)
		entries[i] = LeaderboardEntry{
			UserID:   userID,
			Username: usernames[userID],
			Score:    z.Score,
			Rank:     i + 1,
		}
	}
	return entries, nil
}

func (rs *RedisService) GetUserScore(ctx context.Context, quizCode, userID string) (float64, error) {
	return rs.client.ZScore(ctx, rs.leaderboardKey(quizCode), userID).Result()
}

// TrackAnswerAtomic atomically adds userID to the answered set.
// Returns true if newly added, false if already existed (already answered).
func (rs *RedisService) TrackAnswerAtomic(ctx context.Context, quizCode string, questionIdx int, userID string) (bool, error) {
	key := rs.answeredKey(quizCode, questionIdx)
	added, err := rs.client.SAdd(ctx, key, userID).Result()
	if err != nil {
		return false, err
	}
	rs.client.Expire(ctx, key, keyTTL)
	return added > 0, nil
}

func (rs *RedisService) GetAllScores(ctx context.Context, quizCode string) ([]LeaderboardEntry, error) {
	return rs.GetLeaderboard(ctx, quizCode, -1)
}

func (rs *RedisService) Cleanup(ctx context.Context, quizCode string) error {
	pattern := fmt.Sprintf("quiz:%s:*", quizCode)
	var cursor uint64
	for {
		keys, nextCursor, err := rs.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := rs.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
