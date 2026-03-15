package seed

import (
	"context"
	"log"

	"github.com/chungnguyen/quizz-backend/modules/auth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func seedUsers(db *gorm.DB) {
	repo := auth.NewRepository(db)
	ctx := context.Background()

	users := []struct {
		Username string
		Email    string
		Password string
		Role     string
	}{
		{"admin", "admin@test.com", "123456", auth.RoleAdmin},
		{"player1", "player1@test.com", "123456", auth.RolePlayer},
		{"player2", "player2@test.com", "123456", auth.RolePlayer},
	}

	for _, u := range users {
		existing, _ := repo.GetUserByEmail(ctx, u.Email)
		if existing != nil {
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("seed: failed to hash password for %s: %v", u.Email, err)
			continue
		}

		user := &auth.User{
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: string(hash),
			Role:         u.Role,
		}

		if err := repo.CreateUser(ctx, user); err != nil {
			log.Printf("seed: failed to create user %s: %v", u.Email, err)
			continue
		}

		log.Printf("seed: created user %s (%s)", u.Email, u.Role)
	}
}
