package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/syncraft/backend/internal/security"
)

// RateLimiter returns a Gin middleware that throttles requests based on client IP.
func RateLimiter(limiter *security.IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please wait before retrying",
			})
			return
		}
		c.Next()
	}
}
