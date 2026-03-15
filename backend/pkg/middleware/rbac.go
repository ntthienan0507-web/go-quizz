package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole returns middleware that restricts access to users with one of the given roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		role := c.GetString("role")
		if !allowed[role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "Insufficient permissions",
			})
			return
		}
		c.Next()
	}
}
