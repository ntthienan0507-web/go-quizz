package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (s *Service) Register(ctx context.Context, username, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         RolePlayer,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, *User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", nil, errors.New("invalid credentials")
	}
	if user == nil {
		return "", "", nil, errors.New("invalid credentials")
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); compareErr != nil {
		return "", "", nil, errors.New("invalid credentials")
	}

	accessToken, err := s.generateToken(user)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

func (s *Service) RefreshTokens(ctx context.Context, refreshTokenStr string) (string, string, *User, error) {
	rt, err := s.repo.GetRefreshToken(ctx, refreshTokenStr)
	if err != nil || rt == nil {
		return "", "", nil, errors.New("invalid refresh token")
	}

	if time.Now().After(rt.ExpiresAt) {
		if delErr := s.repo.DeleteRefreshToken(ctx, refreshTokenStr); delErr != nil {
			return "", "", nil, fmt.Errorf("failed to delete expired token: %w", delErr)
		}
		return "", "", nil, errors.New("refresh token expired")
	}

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return "", "", nil, errors.New("user not found")
	}

	accessToken, err := s.generateToken(user)
	if err != nil {
		return "", "", nil, err
	}

	// Atomic rotation: delete old + create new in a single transaction
	newRT, newTokenStr, err := s.buildRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}
	if err := s.repo.RotateRefreshToken(ctx, refreshTokenStr, newRT); err != nil {
		return "", "", nil, fmt.Errorf("failed to rotate token: %w", err)
	}

	return accessToken, newTokenStr, user, nil
}

func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, user *User) error {
	return s.repo.UpdateUser(ctx, user)
}

func (s *Service) generateToken(user *User) (string, error) {
	claims := Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-quizz",
			Audience:  jwt.ClaimStrings{"go-quizz"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) buildRefreshToken(userID uuid.UUID) (*RefreshToken, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, "", err
	}
	tokenStr := hex.EncodeToString(b)

	rt := &RefreshToken{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	return rt, tokenStr, nil
}

func (s *Service) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	rt, tokenStr, err := s.buildRefreshToken(userID)
	if err != nil {
		return "", err
	}
	if err := s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return "", err
	}
	return tokenStr, nil
}
