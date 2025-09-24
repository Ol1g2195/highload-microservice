package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	sm := NewSecurityMiddleware(DefaultSecurityConfig(), logrus.New())
	r.Use(sm.SecurityHeaders())
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("nosniff missing")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("frame deny missing")
	}
	if w.Header().Get("X-XSS-Protection") == "" {
		t.Fatalf("xss header missing")
	}
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := DefaultSecurityConfig()
	cfg.AllowedOrigins = []string{"http://example.com"}
	sm := NewSecurityMiddleware(cfg, logrus.New())
	r.Use(sm.CORS())
	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Fatalf("cors allow origin not set")
	}
}
