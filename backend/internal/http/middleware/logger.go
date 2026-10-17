package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger provides structured HTTP request logging without logging payloads.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// Don't log healthz checks to reduce noise
		if path == "/healthz" || path == "/readyz" {
			return
		}

		log.Printf("[http] %s | %3d | %12v | %s | %s",
			method,
			statusCode,
			latency,
			clientIP,
			path,
		)
	}
}
