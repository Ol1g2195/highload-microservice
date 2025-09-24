package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"highload-microservice/internal/security"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func newSecurityHandler() *SecurityHandler {
	auditor := security.NewSecurityAuditor(logrus.New())
	return NewSecurityHandler(auditor, logrus.New())
}

func TestSecurityHandler_All_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newSecurityHandler()
	r := gin.New()
	r.GET("/security/stats", h.GetSecurityStats)
	r.GET("/security/alerts", h.GetSecurityAlerts)
	r.GET("/security/events", h.GetSecurityEvents)
	r.GET("/security/threats", h.GetThreatIntelligence)
	r.GET("/security/health", h.GetSecurityHealth)

	cases := []string{"/security/stats", "/security/alerts", "/security/events", "/security/threats", "/security/health"}
	for _, p := range cases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", p, nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("path %s expected 200, got %d", p, w.Code)
		}
	}
}
