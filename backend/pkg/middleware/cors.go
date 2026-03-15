package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(origins string) gin.HandlerFunc {
	originList := strings.Split(origins, ",")
	for i := range originList {
		originList[i] = strings.TrimSpace(originList[i])
	}

	// Reject wildcard with credentials
	for _, o := range originList {
		if o == "*" {
			log.Fatal("FATAL: CORS_ORIGINS must not be '*' when AllowCredentials is true")
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     originList,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
