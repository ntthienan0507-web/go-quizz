package question

import "encoding/json"

// --- Request Types ---

type CreateRequest struct {
	Text       string   `json:"text" binding:"required,min=1,max=1000"`
	Options    []string `json:"options" binding:"required,min=2,max=10,dive,required,min=1"`
	CorrectIdx int      `json:"correct_idx" binding:"min=0"`
	Points     int      `json:"points" binding:"omitempty,min=0"`
	OrderNum   int      `json:"order_num" binding:"omitempty,min=0"`
}

// --- Response Types ---

// Response is the admin/owner response that includes the correct answer.
type Response struct {
	ID         string   `json:"id"`
	QuizID     string   `json:"quiz_id"`
	Text       string   `json:"text"`
	Options    []string `json:"options"`
	CorrectIdx int      `json:"correct_idx"`
	Points     int      `json:"points"`
	OrderNum   int      `json:"order_num"`
}

// --- Transformers ---

func ToResponse(q *Question) Response {
	var options []string
	_ = json.Unmarshal(q.Options, &options)
	return Response{
		ID:         q.ID.String(),
		QuizID:     q.QuizID.String(),
		Text:       q.Text,
		Options:    options,
		CorrectIdx: q.CorrectIdx,
		Points:     q.Points,
		OrderNum:   q.OrderNum,
	}
}

func ToResponseList(questions []Question) []Response {
	result := make([]Response, len(questions))
	for i := range questions {
		result[i] = ToResponse(&questions[i])
	}
	return result
}
