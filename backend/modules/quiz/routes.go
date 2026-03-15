package quiz

import (
	"github.com/chungnguyen/quizz-backend/modules/auth"
	"github.com/gin-gonic/gin"
)

type Routes struct {
	controller *Controller
	authSvc    *auth.Service
}

func NewRoutes(controller *Controller, authSvc *auth.Service) *Routes {
	return &Routes{controller: controller, authSvc: authSvc}
}

func (r *Routes) Register(router *gin.RouterGroup, authMiddleware, adminOnly gin.HandlerFunc) *gin.RouterGroup {
	// Protected quiz routes — all require auth, write operations require admin
	quizzes := router.Group("/quizzes")
	quizzes.Use(authMiddleware)
	{
		quizzes.GET("", r.controller.List)
		quizzes.GET("/:id", r.controller.Get)

		// Admin-only write operations
		quizzes.POST("", adminOnly, r.controller.Create)
		quizzes.PUT("/:id", adminOnly, r.controller.Update)
		quizzes.DELETE("/:id", adminOnly, r.controller.Delete)
		quizzes.POST("/:id/start", adminOnly, r.controller.Start)
		quizzes.POST("/:id/finish", adminOnly, r.controller.Finish)
	}

	// Public join route
	router.GET("/quizzes/join/:code", r.controller.Join)

	return quizzes
}
