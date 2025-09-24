package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestRateLimit_AllowsWithinLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewRateLimitMiddleware(RateLimitConfig{Requests: 2, Duration: time.Second}, logrus.New())
	r := gin.New()
	r.Use(mw.RateLimit())
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("unexpected status %d", w.Code)
		}
	}
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw := NewRateLimitMiddleware(RateLimitConfig{Requests: 1, Duration: time.Second}, logrus.New())
	r := gin.New()
	r.Use(mw.RateLimit())
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

	// first ok
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Fatalf("unexpected status %d", w1.Code)
	}

	// second should be 429
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != 429 {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
}
