package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewSecurityMiddleware(DefaultSecurityConfig(), logrus.New())

	r := gin.New()
	r.Use(mw.RequestID())
	r.GET("/ping", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got == "" {
		t.Fatalf("expected X-Request-ID header to be set")
	}
}

func TestRequestID_PropagatesExisting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewSecurityMiddleware(DefaultSecurityConfig(), logrus.New())

	r := gin.New()
	r.Use(mw.RequestID())
	r.GET("/ping", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	req.Header.Set("X-Request-ID", "fixed-id")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "fixed-id" {
		t.Fatalf("expected propagated X-Request-ID, got %q", got)
	}
}
