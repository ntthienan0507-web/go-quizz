package quiz

import (
	"time"

	"github.com/chungnguyen/quizz-backend/modules/quiz/question"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Status Constants ---

const (
	StatusDraft    = "draft"
	StatusActive   = "active"
	StatusFinished = "finished"

	ModeLive      = "live"
	ModeSelfPaced = "self_paced"
)

// --- Quiz ---

type Quiz struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title           string              `gorm:"column:title;size:255;not null"`
	QuizCode        string              `gorm:"column:quiz_code;size:10;uniqueIndex;not null"`
	CreatedBy       uuid.UUID           `gorm:"column:created_by;type:uuid;not null"`
	Status          string              `gorm:"column:status;size:20;default:'draft'"`
	Mode            string              `gorm:"column:mode;size:20;default:'live'"`
	TimePerQuestion int                 `gorm:"column:time_per_question;default:30"`
	Questions       []question.Question `gorm:"foreignKey:QuizID"`
	CreatedAt       time.Time           `gorm:"column:created_at"`
	UpdatedAt       time.Time           `gorm:"column:updated_at"`
}

func (Quiz) TableName() string {
	return "quizzes"
}

func (q *Quiz) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}

// --- QuizResult ---

type QuizResult struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	QuizID     uuid.UUID `gorm:"column:quiz_id;type:uuid;not null;index"`
	UserID     uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	Score      int       `gorm:"column:score;default:0"`
	Rank       int       `gorm:"column:rank"`
	FinishedAt time.Time `gorm:"column:finished_at;autoCreateTime"`
}

func (QuizResult) TableName() string {
	return "quiz_results"
}

func (qr *QuizResult) BeforeCreate(tx *gorm.DB) error {
	if qr.ID == uuid.Nil {
		qr.ID = uuid.New()
	}
	return nil
}

// --- UserAnswer ---

type UserAnswer struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	QuizID      uuid.UUID `gorm:"column:quiz_id;type:uuid;not null"`
	QuestionID  uuid.UUID `gorm:"column:question_id;type:uuid;not null"`
	UserID      uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	SelectedIdx int       `gorm:"column:selected_idx;not null"`
	IsCorrect   bool      `gorm:"column:is_correct;not null"`
	AnsweredAt  time.Time `gorm:"column:answered_at;autoCreateTime"`
}

func (UserAnswer) TableName() string {
	return "user_answers"
}

func (ua *UserAnswer) BeforeCreate(tx *gorm.DB) error {
	if ua.ID == uuid.Nil {
		ua.ID = uuid.New()
	}
	return nil
}
