package quiz

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/chungnguyen/quizz-backend/modules/quiz/question"
	"github.com/chungnguyen/quizz-backend/modules/quiz/realtime"
	"github.com/chungnguyen/quizz-backend/pkg/ws"
	"github.com/google/uuid"
)

type Service struct {
	repo               Repository
	questionRepo       question.Repository
	redisService       *realtime.RedisService
	hubManager         *ws.HubManager
	scorerAdapter      *realtime.ScorerAdapter
	leaderboardAdapter *realtime.LeaderboardAdapter
}

func NewService(
	repo Repository,
	questionRepo question.Repository,
	redisService *realtime.RedisService,
	scoringService *realtime.ScoringService,
	hubManager *ws.HubManager,
) *Service {
	return &Service{
		repo:               repo,
		questionRepo:       questionRepo,
		redisService:       redisService,
		hubManager:         hubManager,
		scorerAdapter:      realtime.NewScorerAdapter(scoringService),
		leaderboardAdapter: realtime.NewLeaderboardAdapter(redisService),
	}
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID) ([]Quiz, error) {
	return s.repo.ListQuizzesByUser(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Quiz, error) {
	return s.repo.GetQuizByID(ctx, id)
}

// GetByIDForOwner returns a quiz only if the caller owns it.
func (s *Service) GetByIDForOwner(ctx context.Context, id, userID uuid.UUID) (*Quiz, error) {
	quiz, err := s.repo.GetQuizByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if quiz == nil {
		return nil, nil
	}
	if quiz.CreatedBy != userID {
		return nil, fmt.Errorf("forbidden")
	}
	return quiz, nil
}

func (s *Service) GetByCode(ctx context.Context, code string) (*Quiz, error) {
	return s.repo.GetQuizByCode(ctx, code)
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, title, mode string, timePerQuestion int) (*Quiz, error) {
	if timePerQuestion == 0 {
		timePerQuestion = 30
	}
	if mode != ModeSelfPaced {
		mode = ModeLive
	}

	quiz := &Quiz{
		Title:           title,
		QuizCode:        generateCode(),
		CreatedBy:       userID,
		Status:          StatusDraft,
		Mode:            mode,
		TimePerQuestion: timePerQuestion,
	}

	if err := s.repo.CreateQuiz(ctx, quiz); err != nil {
		return nil, err
	}
	return quiz, nil
}

func (s *Service) Update(ctx context.Context, id, userID uuid.UUID, title *string, timePerQuestion *int) (*Quiz, error) {
	quiz, err := s.GetByIDForOwner(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if quiz == nil {
		return nil, nil
	}

	if title != nil {
		quiz.Title = *title
	}
	if timePerQuestion != nil && *timePerQuestion > 0 {
		quiz.TimePerQuestion = *timePerQuestion
	}

	if err := s.repo.UpdateQuiz(ctx, quiz); err != nil {
		return nil, err
	}
	return quiz, nil
}

func (s *Service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	quiz, err := s.GetByIDForOwner(ctx, id, userID)
	if err != nil {
		return err
	}
	if quiz == nil {
		return fmt.Errorf("quiz not found")
	}
	return s.repo.DeleteQuiz(ctx, id)
}

func (s *Service) Join(ctx context.Context, code string) (*Quiz, error) {
	quiz, err := s.repo.GetQuizByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if quiz == nil {
		return nil, nil
	}
	if quiz.Status != StatusActive {
		return nil, fmt.Errorf("quiz is not active")
	}
	return quiz, nil
}

func (s *Service) Start(ctx context.Context, id, userID uuid.UUID) (*Quiz, error) {
	quiz, err := s.GetByIDForOwner(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if quiz == nil {
		return nil, nil
	}
	if quiz.Status != StatusDraft {
		return nil, fmt.Errorf("quiz is not in draft status")
	}

	questions, err := s.questionRepo.ListByQuizID(ctx, id)
	if err != nil || len(questions) == 0 {
		return nil, fmt.Errorf("quiz has no questions")
	}

	quiz.Status = StatusActive
	if err := s.repo.UpdateQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	hub := ws.NewHub(
		quiz.QuizCode, quiz.Title, quiz.TimePerQuestion,
		quiz.Mode == ModeSelfPaced,
		quiz.CreatedBy.String(),
		realtime.ToQuestionDataList(questions),
		s.scorerAdapter, s.leaderboardAdapter,
	)
	s.hubManager.CreateHub(hub)

	return quiz, nil
}

func (s *Service) Finish(ctx context.Context, id, userID uuid.UUID) (*Quiz, []ResultResponse, error) {
	quiz, err := s.GetByIDForOwner(ctx, id, userID)
	if err != nil {
		return nil, nil, err
	}
	if quiz == nil {
		return nil, nil, nil
	}

	entries, err := s.redisService.GetAllScores(ctx, quiz.QuizCode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get scores: %w", err)
	}

	var results []QuizResult
	var resultResponses []ResultResponse
	for _, entry := range entries {
		uid, _ := uuid.Parse(entry.UserID)
		results = append(results, QuizResult{
			QuizID: quiz.ID,
			UserID: uid,
			Score:  int(entry.Score),
			Rank:   entry.Rank,
		})
		resultResponses = append(resultResponses, ResultResponse{
			UserID: entry.UserID,
			Score:  int(entry.Score),
			Rank:   entry.Rank,
		})
	}

	if len(results) > 0 {
		if err := s.repo.BatchCreateResults(ctx, results); err != nil {
			return nil, nil, fmt.Errorf("failed to persist results: %w", err)
		}
	}

	quiz.Status = StatusFinished
	_ = s.repo.UpdateQuiz(ctx, quiz)
	_ = s.redisService.Cleanup(ctx, quiz.QuizCode)
	s.hubManager.RemoveHub(quiz.QuizCode)

	return quiz, resultResponses, nil
}

// IsQuizOwner checks if a user owns a quiz (implements question.QuizOwnerChecker).
func (s *Service) IsQuizOwner(ctx context.Context, quizID, userID uuid.UUID) (bool, error) {
	quiz, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return false, err
	}
	if quiz == nil {
		return false, nil
	}
	return quiz.CreatedBy == userID, nil
}

func generateCode() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
