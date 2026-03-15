package player

import (
	"context"

	"github.com/chungnguyen/quizz-backend/modules/auth"
	"github.com/google/uuid"
)

type Service struct {
	repo        IRepository
	authService *auth.Service
}

func NewService(repo IRepository, authService *auth.Service) *Service {
	return &Service{repo: repo, authService: authService}
}

func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID) (*DashboardResponse, error) {
	stats, err := s.repo.GetPlayerStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	globalRank, _ := s.repo.GetPlayerGlobalRank(ctx, userID)

	recent, err := s.repo.GetQuizHistory(ctx, userID, 5, 0)
	if err != nil {
		return nil, err
	}
	if recent == nil {
		recent = []QuizHistoryItem{}
	}

	return &DashboardResponse{
		Stats:      *stats,
		GlobalRank: globalRank,
		Recent:     recent,
	}, nil
}

func (s *Service) GetHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]QuizHistoryItem, int, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	items, err := s.repo.GetQuizHistory(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	if items == nil {
		items = []QuizHistoryItem{}
	}
	return items, page, limit, nil
}

func (s *Service) GetGlobalLeaderboard(ctx context.Context, page, limit int) ([]GlobalLeaderboardEntry, int, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	entries, err := s.repo.GetGlobalLeaderboard(ctx, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	if entries == nil {
		entries = []GlobalLeaderboardEntry{}
	}
	return entries, page, limit, nil
}

type ProfileResponse struct {
	User  auth.UserResponse `json:"user"`
	Stats *PlayerStats      `json:"stats"`
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*ProfileResponse, error) {
	user, err := s.authService.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}

	stats, _ := s.repo.GetPlayerStats(ctx, userID)

	return &ProfileResponse{
		User:  auth.ToUserResponse(user),
		Stats: stats,
	}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, username, email *string) (*auth.UserResponse, error) {
	user, err := s.authService.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, err
	}

	if username != nil && *username != "" {
		user.Username = *username
	}
	if email != nil && *email != "" {
		user.Email = *email
	}

	if err := s.authService.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	resp := auth.ToUserResponse(user)
	return &resp, nil
}
