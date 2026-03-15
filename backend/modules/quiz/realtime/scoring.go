package realtime

import (
	"context"
	"errors"
	"math"

	"github.com/chungnguyen/quizz-backend/modules/quiz/question"
)

var ErrAlreadyAnswered = errors.New("already answered this question")

type ScoringService struct {
	redisService *RedisService
}

func NewScoringService(redisService *RedisService) *ScoringService {
	return &ScoringService{redisService: redisService}
}

type AnswerResult struct {
	IsCorrect     bool    `json:"is_correct"`
	PointsAwarded int     `json:"points_awarded"`
	CorrectIdx    int     `json:"correct_idx"`
	YourTotal     float64 `json:"your_total"`
}

func (ss *ScoringService) ProcessAnswer(
	ctx context.Context,
	quizCode string,
	userID string,
	questionIdx int,
	selectedIdx int,
	q question.Question,
	timeRemaining int,
	timeTotal int,
) (*AnswerResult, error) {
	// Atomic check-and-track: SADD returns 0 if member already existed
	added, err := ss.redisService.TrackAnswerAtomic(ctx, quizCode, questionIdx, userID)
	if err != nil {
		return nil, err
	}
	if !added {
		return nil, ErrAlreadyAnswered
	}

	isCorrect := selectedIdx == q.CorrectIdx
	pointsAwarded := 0

	// Clamp timeRemaining to [0, timeTotal] to prevent score manipulation
	if timeRemaining < 0 {
		timeRemaining = 0
	}
	if timeRemaining > timeTotal {
		timeRemaining = timeTotal
	}

	if isCorrect && timeTotal > 0 {
		ratio := float64(timeRemaining) / float64(timeTotal)
		pointsAwarded = max(int(math.Round(float64(q.Points)*ratio)), 1)

		if err := ss.redisService.UpdateScore(ctx, quizCode, userID, float64(pointsAwarded)); err != nil {
			return nil, err
		}
	}

	total, _ := ss.redisService.GetUserScore(ctx, quizCode, userID)

	return &AnswerResult{
		IsCorrect:     isCorrect,
		PointsAwarded: pointsAwarded,
		CorrectIdx:    q.CorrectIdx,
		YourTotal:     total,
	}, nil
}
