package player

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IRepository interface {
	GetQuizHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]QuizHistoryItem, error)
	GetPlayerStats(ctx context.Context, userID uuid.UUID) (*PlayerStats, error)
	GetGlobalLeaderboard(ctx context.Context, limit, offset int) ([]GlobalLeaderboardEntry, error)
	GetPlayerGlobalRank(ctx context.Context, userID uuid.UUID) (int, error)
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) IRepository {
	return &Repository{db: db}
}

func (r *Repository) GetQuizHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]QuizHistoryItem, error) {
	var items []QuizHistoryItem
	err := r.db.WithContext(ctx).
		Table("quiz_results AS qr").
		Select(`
			qr.quiz_id,
			q.title AS quiz_title,
			q.quiz_code,
			qr.score,
			qr.rank,
			(SELECT COUNT(*) FROM quiz_results WHERE quiz_id = qr.quiz_id) AS total_players,
			qr.finished_at
		`).
		Joins("JOIN quizzes q ON q.id = qr.quiz_id").
		Where("qr.user_id = ?", userID).
		Order("qr.finished_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&items).Error
	return items, err
}

func (r *Repository) GetPlayerStats(ctx context.Context, userID uuid.UUID) (*PlayerStats, error) {
	var stats PlayerStats
	err := r.db.WithContext(ctx).
		Table("quiz_results").
		Select(`
			COUNT(*) AS total_quizzes,
			COALESCE(SUM(score), 0) AS total_score,
			COALESCE(AVG(score), 0) AS avg_score,
			COALESCE(MIN(rank), 0) AS best_rank,
			COUNT(CASE WHEN rank = 1 THEN 1 END) AS wins
		`).
		Where("user_id = ?", userID).
		Scan(&stats).Error
	return &stats, err
}

func (r *Repository) GetGlobalLeaderboard(ctx context.Context, limit, offset int) ([]GlobalLeaderboardEntry, error) {
	var entries []GlobalLeaderboardEntry
	err := r.db.WithContext(ctx).
		Table("quiz_results AS qr").
		Select(`
			qr.user_id,
			u.username,
			COALESCE(SUM(qr.score), 0) AS total_score,
			COUNT(*) AS quizzes_played,
			COALESCE(AVG(qr.score), 0) AS avg_score
		`).
		Joins("INNER JOIN users u ON u.id = qr.user_id").
		Group("qr.user_id, u.username").
		Order("total_score DESC").
		Limit(limit).
		Offset(offset).
		Scan(&entries).Error

	// Assign rank based on position
	for i := range entries {
		entries[i].Rank = offset + i + 1
	}

	return entries, err
}

func (r *Repository) GetPlayerGlobalRank(ctx context.Context, userID uuid.UUID) (int, error) {
	var rank struct {
		Rank int
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT rank FROM (
			SELECT user_id, ROW_NUMBER() OVER (ORDER BY SUM(score) DESC) AS rank
			FROM quiz_results
			WHERE user_id IN (SELECT id FROM users)
			GROUP BY user_id
		) ranked
		WHERE user_id = ?
	`, userID).Scan(&rank).Error

	if err != nil {
		return 0, err
	}
	return rank.Rank, nil
}
