package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
)

// RequestContextMiddleware adds request metadata to context for automatic logging.
// Sets request_id, client_ip, path, and method in context so they're available
// to all downstream handlers and logged automatically.
func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()

		clientIP := c.ClientIP()
		path := c.Request.URL.Path
		method := c.Request.Method

		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logger.ClientIPKey, clientIP)
		ctx = context.WithValue(ctx, logger.PathKey, path)
		ctx = context.WithValue(ctx, logger.MethodKey, method)

		c.Request = c.Request.WithContext(ctx)

		logger.DebugfWithContext(ctx, "request started")

		c.Next()

		// Log request completion with status and latency
		statusCode := c.Writer.Status()
		logger.DebugfWithContext(ctx, "request completed: status=%d", statusCode)
	}
}
