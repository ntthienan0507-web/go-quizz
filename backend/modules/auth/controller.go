package auth

import (
	"net/http"

	"github.com/chungnguyen/quizz-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	refreshCookieName = "refresh_token"
	refreshCookieAge  = 7 * 24 * 60 * 60 // 7 days in seconds
)

func setRefreshCookie(c *gin.Context, token string) {
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, token, refreshCookieAge, "/api/auth", "", secure, true)
}

func clearRefreshCookie(c *gin.Context) {
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, "", -1, "/api/auth", "", secure, true)
}

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// Register creates a new user account
// @Summary Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Registration data"
// @Success 201 {object} response.Response[UserResponse]
// @Failure 400 {object} response.ErrorBody
// @Failure 409 {object} response.ErrorBody
// @Router /auth/register [post]
func (ctrl *Controller) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	user, err := ctrl.service.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		response.Conflict(c, "Username or email already exists")
		return
	}

	response.Created(c, ToUserResponse(user))
}

// Login authenticates a user and returns a JWT token
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response[LoginData]
// @Failure 400 {object} response.ErrorBody
// @Failure 401 {object} response.ErrorBody
// @Router /auth/login [post]
func (ctrl *Controller) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	accessToken, refreshToken, user, err := ctrl.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, "Invalid email or password")
		return
	}

	setRefreshCookie(c, refreshToken)
	response.OK(c, LoginData{
		Token: accessToken,
		User:  ToUserResponse(user),
	})
}

// Refresh exchanges a refresh token for a new access + refresh token pair
func (ctrl *Controller) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	// Read refresh token from httpOnly cookie first, fall back to body
	refreshTokenStr, err := c.Cookie(refreshCookieName)
	if err != nil || refreshTokenStr == "" {
		var req RefreshRequest
		if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
			response.BadRequest(c, "Invalid request body")
			return
		}
		refreshTokenStr = req.RefreshToken
	}

	accessToken, newRefreshToken, user, err := ctrl.service.RefreshTokens(ctx, refreshTokenStr)
	if err != nil {
		clearRefreshCookie(c)
		response.Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	setRefreshCookie(c, newRefreshToken)
	response.OK(c, LoginData{
		Token: accessToken,
		User:  ToUserResponse(user),
	})
}

// Logout clears the refresh token cookie
func (ctrl *Controller) Logout(c *gin.Context) {
	clearRefreshCookie(c)
	response.OK(c, gin.H{"message": "logged out"})
}
