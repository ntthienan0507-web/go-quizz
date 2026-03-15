package auth

import "github.com/gin-gonic/gin"

type Routes struct {
	controller *Controller
}

func NewRoutes(controller *Controller) *Routes {
	return &Routes{controller: controller}
}

func (r *Routes) Register(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	r.RegisterTo(auth)
}

// RegisterTo registers auth handlers on an existing router group.
func (r *Routes) RegisterTo(group *gin.RouterGroup) {
	group.POST("/register", r.controller.Register)
	group.POST("/login", r.controller.Login)
	group.POST("/refresh", r.controller.Refresh)
	group.POST("/logout", r.controller.Logout)
}
