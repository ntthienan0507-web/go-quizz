package player

import "github.com/gin-gonic/gin"

type Routes struct {
	controller *Controller
}

func NewRoutes(controller *Controller) *Routes {
	return &Routes{controller: controller}
}

func (r *Routes) Register(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	player := router.Group("/player")
	player.Use(authMiddleware)
	{
		player.GET("/dashboard", r.controller.GetDashboard)
		player.GET("/history", r.controller.GetHistory)
		player.GET("/leaderboard", r.controller.GetGlobalLeaderboard)
		player.GET("/profile", r.controller.GetProfile)
		player.PUT("/profile", r.controller.UpdateProfile)
	}
}
