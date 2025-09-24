package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"highload-microservice/internal/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func TestSecurityLogging_BasicFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditor := security.NewSecurityAuditor(logrus.New())
	mw := NewSecurityLoggingMiddleware(auditor, logrus.New())

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(mw.LogRequest())
	r.GET("/api/v1/auth/login", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/login", nil)
	req.Header.Set("User-Agent", "curl/7.88")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("unexpected %d", w.Code)
	}
}

func TestSecurityLogging_LogRateLimit429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditor := security.NewSecurityAuditor(logrus.New())
	mw := NewSecurityLoggingMiddleware(auditor, logrus.New())

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(mw.LogRateLimit())
	r.GET("/x", func(c *gin.Context) { c.Status(429) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/x", nil)
	req.Header.Set("User-Agent", "UA")
	r.ServeHTTP(w, req)
	if w.Code != 429 {
		t.Fatalf("want 429, got %d", w.Code)
	}
}

func TestSecurityLogging_LogValidation400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditor := security.NewSecurityAuditor(logrus.New())
	mw := NewSecurityLoggingMiddleware(auditor, logrus.New())

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(mw.LogValidation())
	r.GET("/x", func(c *gin.Context) { c.Set("validation_errors", []string{"e1"}); c.JSON(400, gin.H{"e": "e"}) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/x", nil)
	req.Header.Set("User-Agent", "UA")
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestSecurityLogging_LogAuthorization403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditor := security.NewSecurityAuditor(logrus.New())
	mw := NewSecurityLoggingMiddleware(auditor, logrus.New())

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Set("user_id", uuid.New().String()); c.Next() })
	r.Use(mw.LogAuthorization())
	r.GET("/admin", func(c *gin.Context) { c.Status(403) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("User-Agent", "UA")
	r.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestSecurityLogging_LogAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditor := security.NewSecurityAuditor(logrus.New())
	mw := NewSecurityLoggingMiddleware(auditor, logrus.New())

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Set("user_id", uuid.New().String()); c.Next() })
	r.Use(mw.LogAuthentication())
	r.GET("/auth/me", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("User-Agent", "UA")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestSecurityLogging_LogSuspiciousInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auditor := security.NewSecurityAuditor(logrus.New())
	mw := NewSecurityLoggingMiddleware(auditor, logrus.New())

	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("request_id", "rid"); c.Next() })
	r.Use(mw.LogSuspiciousInput())
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/x", nil)
	req.Header.Set("User-Agent", "sqlmap/1.0")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("want 200, got %d", w.Code)
	}
}
