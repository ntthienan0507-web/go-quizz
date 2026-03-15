package ws

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SubmitAnswerPayload struct {
	QuestionIdx   int `json:"question_idx"`
	SelectedIdx   int `json:"selected_idx"`
	TimeRemaining int `json:"time_remaining"`
}

type WelcomePayload struct {
	QuizTitle      string `json:"quiz_title"`
	TotalQuestions int    `json:"total_questions"`
	Participants   int    `json:"participants"`
	Mode           string `json:"mode"`
}

type PlayerJoinedPayload struct {
	Username         string `json:"username"`
	ParticipantCount int    `json:"participant_count"`
}

type NewQuestionPayload struct {
	QuestionIdx int      `json:"question_idx"`
	Text        string   `json:"text"`
	Options     []string `json:"options"`
	TimeLimit   int      `json:"time_limit"`
}

type AnswerResultPayload struct {
	IsCorrect     bool    `json:"is_correct"`
	PointsAwarded int     `json:"points_awarded"`
	CorrectIdx    int     `json:"correct_idx"`
	YourTotal     float64 `json:"your_total"`
}

type LeaderboardEntry struct {
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	Score    float64 `json:"score"`
	Rank     int     `json:"rank"`
}

type LeaderboardUpdatePayload struct {
	Rankings []LeaderboardEntry `json:"rankings"`
}

type AnswerProgressPayload struct {
	QuestionIdx  int `json:"question_idx"`
	AnsweredCount int `json:"answered_count"`
	TotalPlayers  int `json:"total_players"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

func NewMessage(msgType string, payload interface{}) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	msg := Message{Type: msgType, Payload: p}
	return json.Marshal(msg)
}
