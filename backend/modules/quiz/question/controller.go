package question

import (
	"strings"

	"github.com/chungnguyen/quizz-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func parseUserID(c *gin.Context) (uuid.UUID, bool) {
	uid, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return uuid.Nil, false
	}
	return uid, true
}

func handleServiceError(c *gin.Context, err error, defaultMsg string) {
	if strings.Contains(err.Error(), "forbidden") {
		response.Forbidden(c, "You do not own this quiz")
		return
	}
	response.InternalError(c, defaultMsg)
}

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// List lists all questions for a quiz
// @Summary List questions
// @Tags questions
// @Security BearerAuth
// @Param id path string true "Quiz ID (UUID)"
// @Success 200 {object} response.Response[response.ListData[Response]]
// @Router /quizzes/{id}/questions [get]
func (ctrl *Controller) List(c *gin.Context) {
	ctx := c.Request.Context()
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quiz ID")
		return
	}
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	questions, err := ctrl.service.ListByQuizID(ctx, quizID, userID)
	if err != nil {
		handleServiceError(c, err, "Failed to list questions")
		return
	}

	responses := ToResponseList(questions)
	response.OKList(c, responses, len(responses))
}

// Create adds a question to a quiz
// @Summary Create question
// @Tags questions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID (UUID)"
// @Param body body CreateRequest true "Question data"
// @Success 201 {object} response.Response[Response]
// @Router /quizzes/{id}/questions [post]
func (ctrl *Controller) Create(c *gin.Context) {
	ctx := c.Request.Context()
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quiz ID")
		return
	}
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req CreateRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "Invalid request: "+bindErr.Error())
		return
	}

	q, err := ctrl.service.Create(ctx, quizID, userID, req)
	if err != nil {
		handleServiceError(c, err, "Failed to create question")
		return
	}

	response.Created(c, ToResponse(q))
}

// Update updates a question
// @Summary Update question
// @Tags questions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param qid path string true "Question ID (UUID)"
// @Param body body CreateRequest true "Question data"
// @Success 200 {object} response.Response[Response]
// @Router /questions/{qid} [put]
func (ctrl *Controller) Update(c *gin.Context) {
	ctx := c.Request.Context()
	qid, err := uuid.Parse(c.Param("qid"))
	if err != nil {
		response.BadRequest(c, "Invalid question ID")
		return
	}
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var req CreateRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "Invalid request: "+bindErr.Error())
		return
	}

	q, err := ctrl.service.Update(ctx, qid, userID, req)
	if err != nil {
		handleServiceError(c, err, "Failed to update question")
		return
	}
	if q == nil {
		response.NotFound(c, "Question not found")
		return
	}

	response.OK(c, ToResponse(q))
}

// Delete deletes a question
// @Summary Delete question
// @Tags questions
// @Security BearerAuth
// @Param qid path string true "Question ID (UUID)"
// @Success 200 {object} response.Response[response.MessageData]
// @Router /questions/{qid} [delete]
func (ctrl *Controller) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	qid, err := uuid.Parse(c.Param("qid"))
	if err != nil {
		response.BadRequest(c, "Invalid question ID")
		return
	}
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	if err := ctrl.service.Delete(ctx, qid, userID); err != nil {
		handleServiceError(c, err, "Failed to delete question")
		return
	}

	response.OKMessage(c, "Question deleted")
}
