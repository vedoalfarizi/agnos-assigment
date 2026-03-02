package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vedoalfarizi/hospital-api/internal/logger"
)

// LoggingMiddleware logs incoming requests and outgoing responses with request tracking.
// Masks sensitive data like passwords, tokens, and authorization headers.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Get request ID from context (set by RequestContextMiddleware)
		requestID := ""
		if rid, ok := c.Request.Context().Value(logger.RequestIDKey).(string); ok {
			requestID = rid
		}

		// Log incoming request
		logRequest(c, requestID)

		// Capture response body
		responseBodyWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseBodyWriter

		// Call next handler
		c.Next()

		// Log outgoing response
		duration := time.Since(startTime).Milliseconds()
		logResponse(c, responseBodyWriter, requestID, duration)
	}
}

// responseBodyWriter extracts response body while allowing it to be sent to client
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// logRequest logs incoming request details
func logRequest(c *gin.Context, requestID string) {
	method := c.Request.Method
	path := c.Request.URL.Path
	userAgent := c.Request.UserAgent()
	queryParams := c.Request.URL.RawQuery

	// Log request with essential debugging info
	// Note: client_ip is already in context via RequestContextMiddleware
	logger.InfofWithContext(
		c.Request.Context(),
		"request received: method=%s, path=%s, query=%s, user_agent=%s",
		method,
		path,
		maskQueryParams(queryParams),
		maskUserAgent(userAgent),
	)

	// Log request body for POST/PUT/PATCH requests (excluding file uploads)
	if isBodyLoggingEnabled(method) && c.ContentType() != "application/x-www-form-urlencoded" {
		body, _ := io.ReadAll(c.Request.Body)
		// Restore body for actual handler to read
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		if len(body) > 0 && len(body) < 10000 { // Only log if reasonable size
			maskedBody := maskSensitiveData(string(body))
			logger.DebugfWithContext(
				c.Request.Context(),
				"request body: %s",
				maskedBody,
			)
		}
	}
}

// logResponse logs outgoing response details
func logResponse(c *gin.Context, writer *responseBodyWriter, requestID string, duration int64) {
	statusCode := c.Writer.Status()
	path := c.Request.URL.Path

	// Log response metadata
	logger.InfofWithContext(
		c.Request.Context(),
		"response sent: path=%s, status=%d, duration=%dms",
		path,
		statusCode,
		duration,
	)

	// Log response body only for errors and debugging
	if statusCode >= 400 && writer.body.Len() > 0 && writer.body.Len() < 10000 {
		logger.DebugfWithContext(
			c.Request.Context(),
			"response body: %s",
			maskSensitiveData(writer.body.String()),
		)
	}
}

// isBodyLoggingEnabled checks if request body should be logged
func isBodyLoggingEnabled(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}

// maskSensitiveData masks sensitive fields in request/response bodies
func maskSensitiveData(data string) string {
	sensitiveFields := []string{
		"password",
		"pwd",
		"token",
		"access_token",
		"refresh_token",
		"authorization",
		"secret",
		"api_key",
		"apikey",
		"api-key",
		"auth",
		"bearer",
		"session",
		"sessionid",
		"session_id",
		"csrf",
		"credit_card",
		"creditcard",
		"ssn",
		"social_security",
	}

	result := data
	for _, field := range sensitiveFields {
		// Match patterns like "field":"value" or "field": "value" or field=value
		patterns := []string{
			`"` + field + `"\s*:\s*"[^"]*"`,
			`"` + field + `"\s*:\s*[^,}]*`,
			field + `=[^&]*`,
		}

		for _, pattern := range patterns {
			result = maskPattern(result, pattern, field)
		}
	}

	return result
}

// maskPattern replaces sensitive values matching a pattern
func maskPattern(data, pattern, fieldName string) string {
	// Simple masking regex replacement
	// For JSON fields: replace value with "***"
	if strings.Contains(pattern, `"`) {
		// JSON pattern
		re := buildJSONMaskRegex(fieldName)
		return re.Replace(data, fieldName, "***")
	}
	// Query parameter pattern
	re := buildQueryMaskRegex(fieldName)
	return re.Replace(data, fieldName, "***")
}

// buildJSONMaskRegex creates a regex to mask JSON field values
func buildJSONMaskRegex(fieldName string) *simpleMaskRegex {
	return &simpleMaskRegex{pattern: `"` + fieldName + `"\s*:\s*"[^"]*"`}
}

// buildQueryMaskRegex creates a regex to mask query parameter values
func buildQueryMaskRegex(fieldName string) *simpleMaskRegex {
	return &simpleMaskRegex{pattern: fieldName + `=[^&]*`}
}

// simpleMaskRegex provides basic string masking without full regex
type simpleMaskRegex struct {
	pattern string
}

func (r *simpleMaskRegex) Replace(data, fieldName, replacement string) string {
	result := data

	// Case-insensitive search for JSON patterns
	lowerData := strings.ToLower(data)
	lowerField := strings.ToLower(`"` + fieldName + `"`)

	start := 0
	for {
		idx := strings.Index(lowerData[start:], lowerField)
		if idx == -1 {
			break
		}

		idx += start
		// Find the value part
		colonIdx := strings.Index(result[idx:], ":")
		if colonIdx == -1 {
			break
		}

		colonIdx += idx
		// Find opening quote
		quoteIdx := strings.Index(result[colonIdx:], `"`)
		if quoteIdx == -1 {
			break
		}

		quoteIdx += colonIdx
		// Find closing quote
		endQuoteIdx := strings.Index(result[quoteIdx+1:], `"`)
		if endQuoteIdx == -1 {
			break
		}

		endQuoteIdx += quoteIdx + 1

		// Replace the value
		result = result[:quoteIdx+1] + replacement + result[endQuoteIdx:]
		start = endQuoteIdx + len(replacement) + 1
	}

	return result
}

// maskQueryParams masks sensitive query parameters
func maskQueryParams(query string) string {
	if query == "" {
		return query
	}

	sensitiveParams := map[string]bool{
		"password":      true,
		"token":         true,
		"api_key":       true,
		"secret":        true,
		"authorization": true,
		"access_token":  true,
	}

	params := strings.Split(query, "&")
	for i, param := range params {
		parts := strings.Split(param, "=")
		if len(parts) == 2 {
			key := strings.ToLower(parts[0])
			if sensitiveParams[key] {
				params[i] = parts[0] + "=***"
			}
		}
	}

	return strings.Join(params, "&")
}

// maskUserAgent simplifies user agent for logging (removes some verbose parts)
func maskUserAgent(ua string) string {
	if ua == "" {
		return "unknown"
	}
	// Simplify by truncating very long user agents
	if len(ua) > 100 {
		return ua[:100] + "..."
	}
	return ua
}
