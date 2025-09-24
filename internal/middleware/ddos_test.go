package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestDDoS_BlockAfterThreshold(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ddos := NewDDoSProtection(DDoSConfig{MaxRequests: 1, WindowDuration: time.Second, BlockDuration: time.Second}, logrus.New())
	r := gin.New()
	r.Use(ddos.Protect())
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })

	// first request allowed
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Fatalf("unexpected %d", w1.Code)
	}

	// second within window blocked
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
}
