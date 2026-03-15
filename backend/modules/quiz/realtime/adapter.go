package realtime

import (
	"context"

	"github.com/chungnguyen/quizz-backend/modules/quiz/question"
	"github.com/chungnguyen/quizz-backend/pkg/ws"
)

// --- ScorerAdapter bridges ScoringService to ws.Scorer ---

type ScorerAdapter struct {
	svc *ScoringService
}

func NewScorerAdapter(svc *ScoringService) *ScorerAdapter {
	return &ScorerAdapter{svc: svc}
}

func (a *ScorerAdapter) ProcessAnswer(ctx context.Context, quizCode, userID string, questionIdx, selectedIdx int, q ws.QuestionData, timeRemaining, timeTotal int) (*ws.ScoringResult, error) {
	qModel := question.Question{
		Text:       q.Text,
		Options:    q.Options,
		CorrectIdx: q.CorrectIdx,
		Points:     q.Points,
	}

	result, err := a.svc.ProcessAnswer(ctx, quizCode, userID, questionIdx, selectedIdx, qModel, timeRemaining, timeTotal)
	if err != nil {
		return nil, err
	}

	return &ws.ScoringResult{
		IsCorrect:     result.IsCorrect,
		PointsAwarded: result.PointsAwarded,
		CorrectIdx:    result.CorrectIdx,
		YourTotal:     result.YourTotal,
	}, nil
}

// --- LeaderboardAdapter bridges RedisService to ws.LeaderboardProvider ---

type LeaderboardAdapter struct {
	svc *RedisService
}

func NewLeaderboardAdapter(svc *RedisService) *LeaderboardAdapter {
	return &LeaderboardAdapter{svc: svc}
}

func (a *LeaderboardAdapter) GetLeaderboard(ctx context.Context, quizCode string, top int) ([]ws.LeaderboardEntryData, error) {
	entries, err := a.svc.GetLeaderboard(ctx, quizCode, top)
	if err != nil {
		return nil, err
	}

	result := make([]ws.LeaderboardEntryData, len(entries))
	for i, e := range entries {
		result[i] = ws.LeaderboardEntryData{
			UserID:   e.UserID,
			Username: e.Username,
			Score:    e.Score,
			Rank:     e.Rank,
		}
	}
	return result, nil
}

func (a *LeaderboardAdapter) SetUsername(ctx context.Context, quizCode, userID, username string) error {
	return a.svc.SetUsername(ctx, quizCode, userID, username)
}

// ToQuestionDataList converts question models to ws.QuestionData.
func ToQuestionDataList(questions []question.Question) []ws.QuestionData {
	result := make([]ws.QuestionData, len(questions))
	for i, q := range questions {
		result[i] = ws.QuestionData{
			Text:       q.Text,
			Options:    q.Options,
			CorrectIdx: q.CorrectIdx,
			Points:     q.Points,
		}
	}
	return result
}
