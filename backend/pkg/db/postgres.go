package db

import (
	"log"

	"github.com/chungnguyen/quizz-backend/modules/auth"
	"github.com/chungnguyen/quizz-backend/modules/quiz"
	"github.com/chungnguyen/quizz-backend/modules/quiz/question"
	"github.com/chungnguyen/quizz-backend/pkg/config"
	"github.com/chungnguyen/quizz-backend/pkg/db/seed"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(cfg *config.Config) *gorm.DB {
	dsn := cfg.DSN()
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&auth.User{},
		&auth.RefreshToken{},
		&quiz.Quiz{},
		&question.Question{},
		&quiz.QuizResult{},
		&quiz.UserAnswer{},
	); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	log.Println("database connected and migrated")

	seed.Run(db)

	return db
}
