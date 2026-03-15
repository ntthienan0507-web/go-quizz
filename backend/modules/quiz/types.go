package quiz

import "time"

// --- Request Types ---

type CreateQuizRequest struct {
	Title           string `json:"title" binding:"required,min=1,max=200"`
	Mode            string `json:"mode" binding:"omitempty,oneof=live self_paced"`
	TimePerQuestion int    `json:"time_per_question" binding:"omitempty,min=5,max=300"`
}

type UpdateQuizRequest struct {
	Title           *string `json:"title" binding:"omitempty,min=1,max=200"`
	TimePerQuestion *int    `json:"time_per_question" binding:"omitempty,min=5,max=300"`
}

// --- Response Types ---

type QuizResponse struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	QuizCode        string    `json:"quiz_code"`
	CreatedBy       string    `json:"created_by"`
	Status          string    `json:"status"`
	Mode            string    `json:"mode"`
	TimePerQuestion int       `json:"time_per_question"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type QuizJoinData struct {
	QuizID   string `json:"quiz_id"`
	Title    string `json:"title"`
	QuizCode string `json:"quiz_code"`
	Status   string `json:"status"`
	Mode     string `json:"mode"`
}

type QuizStartData struct {
	Message  string `json:"message"`
	QuizCode string `json:"quiz_code"`
}

type QuizFinishData struct {
	Message string           `json:"message"`
	Results []ResultResponse `json:"results"`
}

type ResultResponse struct {
	UserID string `json:"user_id"`
	Score  int    `json:"score"`
	Rank   int    `json:"rank"`
}

// --- Transformers ---

func ToQuizResponse(q *Quiz) QuizResponse {
	return QuizResponse{
		ID:              q.ID.String(),
		Title:           q.Title,
		QuizCode:        q.QuizCode,
		CreatedBy:       q.CreatedBy.String(),
		Status:          q.Status,
		Mode:            q.Mode,
		TimePerQuestion: q.TimePerQuestion,
		CreatedAt:       q.CreatedAt,
		UpdatedAt:       q.UpdatedAt,
	}
}

func ToQuizResponseList(quizzes []Quiz) []QuizResponse {
	result := make([]QuizResponse, len(quizzes))
	for i := range quizzes {
		result[i] = ToQuizResponse(&quizzes[i])
	}
	return result
}
