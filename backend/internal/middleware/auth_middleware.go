// Package middleware holds cross-cutting HTTP concerns (auth, RBAC, ...).
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ymmo/internal/security"
)

// Context keys used to share the authenticated identity with handlers.
const (
	ctxUserID = "currentUserID"
	ctxRole   = "currentRole"
)

// Auth validates the Bearer JWT and stores the identity in the context.
// Any request without a valid token is rejected with 401.
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := security.ParseToken(jwtSecret, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// RequireRole guards a route so only the given roles may proceed.
// Used later by internal endpoints (agent, director, HQ...).
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		current := CurrentRole(c)
		for _, allowed := range roles {
			if allowed == current {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// CurrentUserID returns the authenticated user's ID (0 if none).
func CurrentUserID(c *gin.Context) uint {
	if v, ok := c.Get(ctxUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// CurrentRole returns the authenticated user's role code ("" if none).
func CurrentRole(c *gin.Context) string {
	if v, ok := c.Get(ctxRole); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
