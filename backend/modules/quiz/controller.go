package quiz

import (
	"github.com/chungnguyen/quizz-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// List returns all quizzes for the authenticated user
// @Summary List quizzes
// @Tags quizzes
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response[response.ListData[QuizResponse]]
// @Router /quizzes [get]
func (ctrl *Controller) List(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	quizzes, err := ctrl.service.ListByUser(ctx, userID)
	if err != nil {
		response.InternalError(c, "Failed to list quizzes")
		return
	}

	responses := ToQuizResponseList(quizzes)
	response.OKList(c, responses, len(responses))
}

// Create creates a new quiz
// @Summary Create quiz
// @Tags quizzes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateQuizRequest true "Quiz data"
// @Success 201 {object} response.Response[QuizResponse]
// @Router /quizzes [post]
func (ctrl *Controller) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	quiz, err := ctrl.service.Create(ctx, userID, req.Title, req.Mode, req.TimePerQuestion)
	if err != nil {
		response.InternalError(c, "Failed to create quiz")
		return
	}

	response.Created(c, ToQuizResponse(quiz))
}

// Get returns a quiz by ID
// @Summary Get quiz by ID
// @Tags quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID (UUID)"
// @Success 200 {object} response.Response[QuizResponse]
// @Failure 404 {object} response.ErrorBody
// @Router /quizzes/{id} [get]
func (ctrl *Controller) Get(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quiz ID")
		return
	}
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	quiz, err := ctrl.service.GetByIDForOwner(ctx, id, userID)
	if err != nil {
		response.Forbidden(c, "You do not own this quiz")
		return
	}
	if quiz == nil {
		response.NotFound(c, "Quiz not found")
		return
	}

	response.OK(c, ToQuizResponse(quiz))
}

// Update updates a quiz
// @Summary Update quiz
// @Tags quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID (UUID)"
// @Param body body UpdateQuizRequest true "Quiz update data"
// @Success 200 {object} response.Response[QuizResponse]
// @Router /quizzes/{id} [put]
func (ctrl *Controller) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quiz ID")
		return
	}
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateQuizRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "Invalid request: "+bindErr.Error())
		return
	}

	quiz, err := ctrl.service.Update(ctx, id, userID, req.Title, req.TimePerQuestion)
	if err != nil {
		response.InternalError(c, "Failed to update quiz")
		return
	}
	if quiz == nil {
		response.NotFound(c, "Quiz not found")
		return
	}

	response.OK(c, ToQuizResponse(quiz))
}

// Delete deletes a quiz
// @Summary Delete quiz
// @Tags quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID (UUID)"
// @Success 200 {object} response.Response[response.MessageData]
// @Router /quizzes/{id} [delete]
func (ctrl *Controller) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quiz ID")
		return
	}
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := ctrl.service.Delete(ctx, id, userID); err != nil {
		response.InternalError(c, "Failed to delete quiz")
		return
	}

	response.OKMessage(c, "Quiz deleted")
}

// Join validates a quiz code for joining
// @Summary Join quiz by code
// @Tags quizzes
// @Param code path string true "Quiz code"
// @Success 200 {object} response.Response[QuizJoinData]
// @Router /quizzes/join/{code} [get]
func (ctrl *Controller) Join(c *gin.Context) {
	ctx := c.Request.Context()
	code := c.Param("code")

	quiz, err := ctrl.service.Join(ctx, code)
	if err != nil {
		response.BadRequest(c, "Quiz is not available for joining")
		return
	}
	if quiz == nil {
		response.NotFound(c, "Quiz not found")
		return
	}

	response.OK(c, QuizJoinData{
		QuizID:   quiz.ID.String(),
		Title:    quiz.Title,
		QuizCode: quiz.QuizCode,
		Status:   quiz.Status,
		Mode:     quiz.Mode,
	})
}

// Start activates a quiz and creates the WebSocket hub
// @Summary Start quiz
// @Tags quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID (UUID)"
// @Success 200 {object} response.Response[QuizStartData]
// @Router /quizzes/{id}/start [post]
func (ctrl *Controller) Start(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quiz ID")
		return
	}
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	quiz, err := ctrl.service.Start(ctx, id, userID)
	if err != nil {
		response.BadRequest(c, "Cannot start quiz: ensure it is in draft status and has questions")
		return
	}
	if quiz == nil {
		response.NotFound(c, "Quiz not found")
		return
	}

	response.OK(c, QuizStartData{
		Message:  "Quiz started",
		QuizCode: quiz.QuizCode,
	})
}

// Finish ends a quiz and persists results from Redis to PostgreSQL
// @Summary Finish quiz
// @Tags quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID (UUID)"
// @Success 200 {object} response.Response[QuizFinishData]
// @Router /quizzes/{id}/finish [post]
func (ctrl *Controller) Finish(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quiz ID")
		return
	}
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	quiz, resultResponses, err := ctrl.service.Finish(ctx, id, userID)
	if err != nil {
		response.InternalError(c, "Failed to finish quiz")
		return
	}
	if quiz == nil {
		response.NotFound(c, "Quiz not found")
		return
	}

	response.OK(c, QuizFinishData{
		Message: "Quiz finished",
		Results: resultResponses,
	})
}
