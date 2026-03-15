package question

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Question struct {
	ID         uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	QuizID     uuid.UUID       `gorm:"column:quiz_id;type:uuid;not null;index"`
	Text       string          `gorm:"column:text;type:text;not null"`
	Options    json.RawMessage `gorm:"column:options;type:jsonb;not null"`
	CorrectIdx int             `gorm:"column:correct_idx;not null"`
	Points     int             `gorm:"column:points;default:10"`
	OrderNum   int             `gorm:"column:order_num;not null"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
}

func (Question) TableName() string {
	return "questions"
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}
