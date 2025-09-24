package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type createReq struct {
	Email string `json:"email" validate:"required,email,no_sql_injection,no_xss"`
}

func TestValidationMiddleware_ValidateRequest_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	vm := NewValidationMiddleware(logrus.New())
	r.POST("/", vm.ValidateRequest(&createReq{}), func(c *gin.Context) { c.String(200, "ok") })

	body := bytes.NewBufferString(`{"email":"user@example.com"}`)
	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
}

func TestValidationMiddleware_ValidateRequest_Bad(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	vm := NewValidationMiddleware(logrus.New())
	r.POST("/", vm.ValidateRequest(&createReq{}), func(c *gin.Context) { c.String(200, "ok") })

	body := bytes.NewBufferString(`{"email":"bad@<script>.com"}`)
	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
