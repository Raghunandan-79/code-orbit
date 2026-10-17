package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SafeRecovery recovers from panics and returns a generic safe JSON response.
func SafeRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[panic recovered] %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "an unexpected internal error occurred",
				})
			}
		}()
		c.Next()
	}
}
