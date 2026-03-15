package question

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// QuizOwnerChecker verifies quiz ownership.
type QuizOwnerChecker interface {
	IsQuizOwner(ctx context.Context, quizID, userID uuid.UUID) (bool, error)
}

type Service struct {
	repo       Repository
	quizOwner  QuizOwnerChecker
}

func NewService(repo Repository, quizOwner QuizOwnerChecker) *Service {
	return &Service{repo: repo, quizOwner: quizOwner}
}

func (s *Service) checkOwnership(ctx context.Context, quizID, userID uuid.UUID) error {
	ok, err := s.quizOwner.IsQuizOwner(ctx, quizID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("forbidden")
	}
	return nil
}

func (s *Service) ListByQuizID(ctx context.Context, quizID, userID uuid.UUID) ([]Question, error) {
	if err := s.checkOwnership(ctx, quizID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByQuizID(ctx, quizID)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Question, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, quizID, userID uuid.UUID, req CreateRequest) (*Question, error) {
	if err := s.checkOwnership(ctx, quizID, userID); err != nil {
		return nil, err
	}

	optionsJSON, _ := json.Marshal(req.Options)
	points := req.Points
	if points == 0 {
		points = 10
	}

	q := &Question{
		QuizID:     quizID,
		Text:       req.Text,
		Options:    optionsJSON,
		CorrectIdx: req.CorrectIdx,
		Points:     points,
		OrderNum:   req.OrderNum,
	}

	if err := s.repo.Create(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *Service) Update(ctx context.Context, id, userID uuid.UUID, req CreateRequest) (*Question, error) {
	q, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, nil
	}
	if err := s.checkOwnership(ctx, q.QuizID, userID); err != nil {
		return nil, err
	}

	optionsJSON, _ := json.Marshal(req.Options)
	q.Text = req.Text
	q.Options = optionsJSON
	q.CorrectIdx = req.CorrectIdx
	if req.Points > 0 {
		q.Points = req.Points
	}
	q.OrderNum = req.OrderNum

	if err := s.repo.Update(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *Service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	q, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if q == nil {
		return fmt.Errorf("question not found")
	}
	if err := s.checkOwnership(ctx, q.QuizID, userID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
