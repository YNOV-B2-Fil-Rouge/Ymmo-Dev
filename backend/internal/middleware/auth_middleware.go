// Package middleware holds cross-cutting HTTP concerns (auth, RBAC, ...).
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ymmo/internal/security"
)

const (
	ctxUserID = "currentUserID"
	ctxRole   = "currentRole"
)

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

func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header != "" && strings.HasPrefix(header, "Bearer ") {
			tokenString := strings.TrimPrefix(header, "Bearer ")
			if claims, err := security.ParseToken(jwtSecret, tokenString); err == nil {
				c.Set(ctxUserID, claims.UserID)
				c.Set(ctxRole, claims.Role)
			}
		}
		c.Next()
	}
}

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

func CurrentUserID(c *gin.Context) uint {
	if v, ok := c.Get(ctxUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

func CurrentRole(c *gin.Context) string {
	if v, ok := c.Get(ctxRole); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
