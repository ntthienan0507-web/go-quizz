package quiz

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateQuiz(ctx context.Context, quiz *Quiz) error
	GetQuizByID(ctx context.Context, id uuid.UUID) (*Quiz, error)
	GetQuizByCode(ctx context.Context, code string) (*Quiz, error)
	ListQuizzesByUser(ctx context.Context, userID uuid.UUID) ([]Quiz, error)
	UpdateQuiz(ctx context.Context, quiz *Quiz) error
	DeleteQuiz(ctx context.Context, id uuid.UUID) error
	BatchCreateResults(ctx context.Context, results []QuizResult) error
	GetResultsByQuizID(ctx context.Context, quizID uuid.UUID) ([]QuizResult, error)
	BatchCreateAnswers(ctx context.Context, answers []UserAnswer) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// --- Quiz CRUD ---

func (r *GormRepository) CreateQuiz(ctx context.Context, quiz *Quiz) error {
	return r.db.WithContext(ctx).Create(quiz).Error
}

func (r *GormRepository) GetQuizByID(ctx context.Context, id uuid.UUID) (*Quiz, error) {
	var quiz Quiz
	err := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_num ASC")
		}).
		First(&quiz, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &quiz, nil
}

func (r *GormRepository) GetQuizByCode(ctx context.Context, code string) (*Quiz, error) {
	var quiz Quiz
	err := r.db.WithContext(ctx).Where("quiz_code = ?", code).First(&quiz).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &quiz, nil
}

func (r *GormRepository) ListQuizzesByUser(ctx context.Context, userID uuid.UUID) ([]Quiz, error) {
	var quizzes []Quiz
	err := r.db.WithContext(ctx).
		Where("created_by = ?", userID).
		Order("created_at DESC").
		Find(&quizzes).Error
	return quizzes, err
}

func (r *GormRepository) UpdateQuiz(ctx context.Context, quiz *Quiz) error {
	return r.db.WithContext(ctx).Model(quiz).Updates(map[string]any{
		"title":             quiz.Title,
		"status":            quiz.Status,
		"time_per_question": quiz.TimePerQuestion,
	}).Error
}

func (r *GormRepository) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Quiz{}, "id = ?", id).Error
}

// --- Result ---

func (r *GormRepository) BatchCreateResults(ctx context.Context, results []QuizResult) error {
	return r.db.WithContext(ctx).Create(&results).Error
}

func (r *GormRepository) GetResultsByQuizID(ctx context.Context, quizID uuid.UUID) ([]QuizResult, error) {
	var results []QuizResult
	err := r.db.WithContext(ctx).
		Where("quiz_id = ?", quizID).
		Order("rank ASC").
		Find(&results).Error
	return results, err
}

func (r *GormRepository) BatchCreateAnswers(ctx context.Context, answers []UserAnswer) error {
	return r.db.WithContext(ctx).Create(&answers).Error
}
