package question

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, question *Question) error
	ListByQuizID(ctx context.Context, quizID uuid.UUID) ([]Question, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Question, error)
	Update(ctx context.Context, question *Question) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, question *Question) error {
	return r.db.WithContext(ctx).Create(question).Error
}

func (r *GormRepository) ListByQuizID(ctx context.Context, quizID uuid.UUID) ([]Question, error) {
	var questions []Question
	err := r.db.WithContext(ctx).
		Where("quiz_id = ?", quizID).
		Order("order_num ASC").
		Find(&questions).Error
	return questions, err
}

func (r *GormRepository) GetByID(ctx context.Context, id uuid.UUID) (*Question, error) {
	var question Question
	err := r.db.WithContext(ctx).First(&question, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &question, nil
}

func (r *GormRepository) Update(ctx context.Context, question *Question) error {
	return r.db.WithContext(ctx).Save(question).Error
}

func (r *GormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Question{}, "id = ?", id).Error
}
