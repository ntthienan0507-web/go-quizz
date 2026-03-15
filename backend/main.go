package main

import (
	"log"
	"time"

	"github.com/chungnguyen/quizz-backend/pkg/config"
	"github.com/chungnguyen/quizz-backend/pkg/db"
	"github.com/chungnguyen/quizz-backend/pkg/middleware"
	"github.com/chungnguyen/quizz-backend/pkg/ws"

	authmod "github.com/chungnguyen/quizz-backend/modules/auth"
	playermod "github.com/chungnguyen/quizz-backend/modules/player"
	quizmod "github.com/chungnguyen/quizz-backend/modules/quiz"
	questionmod "github.com/chungnguyen/quizz-backend/modules/quiz/question"
	"github.com/chungnguyen/quizz-backend/modules/quiz/realtime"

	_ "github.com/chungnguyen/quizz-backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Quiz Platform API
// @version 1.0
// @description Real-time quiz platform with live leaderboards.

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}"

func main() {
	cfg := config.Load()

	pgDB := db.NewPostgres(cfg)
	redisClient := db.NewRedis(cfg.RedisAddr, cfg.RedisURL)

	// --- Auth module ---
	authRepo := authmod.NewRepository(pgDB)
	authService := authmod.NewService(authRepo, cfg.JWTSecret)
	authController := authmod.NewController(authService)
	authRoutes := authmod.NewRoutes(authController)

	// --- Question module (repo first, service wired after quiz) ---
	questionRepo := questionmod.NewRepository(pgDB)

	// --- Realtime services ---
	redisService := realtime.NewRedisService(redisClient)
	scoringService := realtime.NewScoringService(redisService)

	// --- Quiz module ---
	quizRepo := quizmod.NewRepository(pgDB)
	hubManager := ws.NewHubManager()
	quizService := quizmod.NewService(quizRepo, questionRepo, redisService, scoringService, hubManager)

	// --- Question service (needs quizService for ownership check) ---
	questionService := questionmod.NewService(questionRepo, quizService)

	questionController := questionmod.NewController(questionService)
	questionRoutes := questionmod.NewRoutes(questionController)
	quizController := quizmod.NewController(quizService)
	quizRoutes := quizmod.NewRoutes(quizController, authService)

	// --- Player module ---
	playerRepo := playermod.NewRepository(pgDB)
	playerService := playermod.NewService(playerRepo, authService)
	playerController := playermod.NewController(playerService)
	playerRoutes := playermod.NewRoutes(playerController)

	// --- WebSocket handler ---
	leaderboardAdapter := realtime.NewLeaderboardAdapter(redisService)
	wsHandler := ws.NewWSHandler(hubManager, authService, leaderboardAdapter, cfg.CORSOrigins)

	// --- Rate limiters ---
	authRL := middleware.RateLimit(redisClient, middleware.RateLimitConfig{
		Max: 10, Window: 1 * time.Minute, // 10 login/register attempts per minute
	})
	apiRL := middleware.RateLimit(redisClient, middleware.RateLimitConfig{
		Max: 60, Window: 1 * time.Minute, // 60 requests per minute for general API
	})
	wsRL := middleware.RateLimit(redisClient, middleware.RateLimitConfig{
		Max: 5, Window: 1 * time.Minute, // 5 WS connections per minute per IP
	})

	// --- Router ---
	r := gin.Default()
	_ = r.SetTrustedProxies(nil) // trust no proxy by default; set to actual proxy CIDRs in production
	r.MaxMultipartMemory = 2 << 20 // 2 MB
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(cfg.CORSOrigins))

	// Swagger UI — only in non-release mode
	if gin.Mode() != gin.ReleaseMode {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := r.Group("/api")
	api.Use(apiRL)
	{
		authMiddleware := middleware.Auth(authService)

		// Auth routes — stricter rate limit
		auth := api.Group("/auth")
		auth.Use(authRL)
		authRoutes.RegisterTo(auth)

		adminOnly := middleware.RequireRole("admin")

		quizGroup := quizRoutes.Register(api, authMiddleware, adminOnly)

		// Question routes (nested under quizzes + standalone) — admin only for writes
		questions := api.Group("/questions")
		questions.Use(authMiddleware, adminOnly)
		questionRoutes.Register(quizGroup, questions, adminOnly)

		// Player routes
		playerRoutes.Register(api, authMiddleware)
	}

	r.GET("/ws/:code", wsRL, wsHandler.Handle)

	log.Printf("server starting on :%s", cfg.Port)
	log.Printf("swagger UI: http://localhost:%s/swagger/index.html", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
