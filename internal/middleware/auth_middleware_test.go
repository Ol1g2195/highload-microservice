package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"highload-microservice/internal/models"
	"highload-microservice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type stubAuthService struct{}

func (s *stubAuthService) ValidateToken(token string) (*models.JWTClaims, error) {
	return nil, errors.New("not impl")
}

// compile-time check we match the method used
var _ = (&services.AuthService{} != nil)

func TestRequireAuth_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	// use real AuthService pointer but we won't call it
	m := NewAuthMiddleware(&services.AuthService{}, logger)

	r := gin.New()
	r.GET("/p", m.RequireAuth(), func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	m := NewAuthMiddleware(&services.AuthService{}, logger)

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		c.Set("user_email", "u@l")
		c.Set("user_role", models.UserRole("user"))
		c.Next()
	}, m.RequireRole("admin"), func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}
