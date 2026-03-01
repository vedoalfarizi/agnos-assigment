package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/vedoalfarizi/hospital-api/internal/handler"
)

// Claims represents the JWT claims structure matching StaffService token generation
type Claims struct {
	StaffID    int `json:"staff_id"`
	HospitalID int `json:"hospital_id"`
	jwt.StandardClaims
}

// AuthMiddleware returns a Gin middleware that validates JWT tokens from the Authorization header.
// On success, it extracts hospital_id and staff_id and injects them into the request context.
// Returns 401 if token is missing, invalid, or expired.
func AuthMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			handler.Error(c, 401, "UNAUTHORIZED", "Authentication required or token invalid")
			c.Abort()
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			handler.Error(c, 401, "UNAUTHORIZED", "Authentication required or token invalid")
			c.Abort()
			return
		}

		// Extract token string
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse token with claims
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil || !token.Valid {
			handler.Error(c, 401, "UNAUTHORIZED", "Authentication required or token invalid")
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			handler.Error(c, 401, "UNAUTHORIZED", "Authentication required or token invalid")
			c.Abort()
			return
		}

		// Inject claims into request context
		c.Set("hospital_id", claims.HospitalID)
		c.Set("staff_id", claims.StaffID)

		c.Next()
	}
}
