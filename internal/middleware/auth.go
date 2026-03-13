package middleware

import (
	"net/http"
	"strings"

	"github.com/Bimidu/ctse-order-service/internal/config"
	"github.com/Bimidu/ctse-order-service/grpc/clients"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// AuthRequired verifie the JWT. It first tries local verification for performance,
// then falls back to the Auth Service gRPC call if local verification fails.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			return
		}

		claims, err := verifyLocal(token)
		if err == nil {
			c.Set("user_id", claims.ID)
			c.Set("role", claims.Role)
			c.Next()
			return
		}

		// Fall back to Auth Service gRPC verification
		if clients.Auth != nil {
			resp, grpcErr := clients.Auth.VerifyToken(token)
			if grpcErr == nil && resp.Valid {
				c.Set("user_id", resp.UserId)
				c.Set("role", resp.Role)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
	}
}

// RoleRequired restricts access to users with a specific role.
func RoleRequired(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole.(string) != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func verifyLocal(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.App.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

// GetUserID extracts the authenticated user ID from gin context.
func GetUserID(c *gin.Context) string {
	id, _ := c.Get("user_id")
	return id.(string)
}

// GetUserRole extracts the authenticated user role from gin context.
func GetUserRole(c *gin.Context) string {
	role, _ := c.Get("role")
	return role.(string)
}
