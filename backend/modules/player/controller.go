package player

import (
	"strconv"

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

func (ctrl *Controller) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	dashboard, err := ctrl.service.GetDashboard(ctx, userID)
	if err != nil {
		response.InternalError(c, "Failed to get dashboard")
		return
	}

	response.OK(c, dashboard)
}

func (ctrl *Controller) GetHistory(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	items, pg, lim, err := ctrl.service.GetHistory(ctx, userID, page, limit)
	if err != nil {
		response.InternalError(c, "Failed to get history")
		return
	}

	response.OK(c, gin.H{
		"items": items,
		"page":  pg,
		"limit": lim,
	})
}

func (ctrl *Controller) GetGlobalLeaderboard(c *gin.Context) {
	ctx := c.Request.Context()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	entries, pg, lim, err := ctrl.service.GetGlobalLeaderboard(ctx, page, limit)
	if err != nil {
		response.InternalError(c, "Failed to get leaderboard")
		return
	}

	response.OK(c, gin.H{
		"items": entries,
		"page":  pg,
		"limit": lim,
	})
}

func (ctrl *Controller) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	profile, err := ctrl.service.GetProfile(ctx, userID)
	if err != nil {
		response.InternalError(c, "Failed to get profile")
		return
	}
	if profile == nil {
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, profile)
}

func (ctrl *Controller) UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateProfileRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	userResp, err := ctrl.service.UpdateProfile(ctx, userID, req.Username, req.Email)
	if err != nil {
		response.Conflict(c, "Username or email already taken")
		return
	}
	if userResp == nil {
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, userResp)
}
