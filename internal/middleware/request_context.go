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
		// Generate unique request ID
		requestID := uuid.New().String()

		// Extract request metadata
		clientIP := c.ClientIP()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Create new context with request metadata
		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logger.ClientIPKey, clientIP)
		ctx = context.WithValue(ctx, logger.PathKey, path)
		ctx = context.WithValue(ctx, logger.MethodKey, method)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Log request start
		logger.DebugfWithContext(ctx, "request started")

		// Call next handler
		c.Next()

		// Log request completion with status and latency
		statusCode := c.Writer.Status()
		logger.DebugfWithContext(ctx, "request completed: status=%d", statusCode)
	}
}
