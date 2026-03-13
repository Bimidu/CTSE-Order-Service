package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bimidu/ctse-order-service/internal/config"
	"github.com/Bimidu/ctse-order-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthRequiredMiddlewareValidToken(t *testing.T) {
	config.App = &config.Config{JWTSecret: "test-secret"}
	gin.SetMode(gin.TestMode)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   "user-1",
		"role": "user",
	})
	signed, err := token.SignedString([]byte(config.App.JWTSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	r := gin.New()
	r.Use(middleware.AuthRequired())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func TestRoleRequiredMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	r.Use(middleware.RoleRequired("admin"))
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

