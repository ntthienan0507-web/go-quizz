package question

import "github.com/gin-gonic/gin"

type Routes struct {
	controller *Controller
}

func NewRoutes(controller *Controller) *Routes {
	return &Routes{controller: controller}
}

func (r *Routes) Register(quizGroup *gin.RouterGroup, questionGroup *gin.RouterGroup, adminOnly ...gin.HandlerFunc) {
	// Nested under /quizzes/:id — auth applied via quizGroup
	quizGroup.GET("/:id/questions", r.controller.List)
	if len(adminOnly) > 0 {
		quizGroup.POST("/:id/questions", adminOnly[0], r.controller.Create)
	} else {
		quizGroup.POST("/:id/questions", r.controller.Create)
	}

	// Standalone /questions/:qid — adminOnly applied at group level
	questionGroup.PUT("/:qid", r.controller.Update)
	questionGroup.DELETE("/:qid", r.controller.Delete)
}
