package player

import "time"

type QuizHistoryItem struct {
	QuizID       string    `json:"quiz_id"`
	QuizTitle    string    `json:"quiz_title"`
	QuizCode     string    `json:"quiz_code"`
	Score        int       `json:"score"`
	Rank         int       `json:"rank"`
	TotalPlayers int       `json:"total_players"`
	FinishedAt   time.Time `json:"finished_at"`
}

type PlayerStats struct {
	TotalQuizzes int     `json:"total_quizzes"`
	TotalScore   int     `json:"total_score"`
	AvgScore     float64 `json:"avg_score"`
	BestRank     int     `json:"best_rank"`
	Wins         int     `json:"wins"`
}

type GlobalLeaderboardEntry struct {
	UserID       string  `json:"user_id"`
	Username     string  `json:"username"`
	TotalScore   int     `json:"total_score"`
	QuizzesPlayed int    `json:"quizzes_played"`
	AvgScore     float64 `json:"avg_score"`
	Rank         int     `json:"rank"`
}

type DashboardResponse struct {
	Stats      PlayerStats       `json:"stats"`
	GlobalRank int               `json:"global_rank"`
	Recent     []QuizHistoryItem `json:"recent"`
}

type UpdateProfileRequest struct {
	Username *string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    *string `json:"email" binding:"omitempty,email"`
}
