package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"highload-microservice/internal/models"
	"highload-microservice/internal/security"
	"highload-microservice/internal/services"

	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func newAuthHandlerForTest(t *testing.T) (*AuthHandler, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	logger := logrus.New()
	cfg := services.AuthConfig{JWTSecret: "secret", JWTExpiration: time.Hour, RefreshExpiration: 24 * time.Hour, APIKeyLength: 4}
	authSvc := services.NewAuthService(db, logger, cfg)
	auditor := security.NewSecurityAuditor(logger)
	h := NewAuthHandler(authSvc, auditor, logger)
	cleanup := func() { _ = db.Close() }
	return h, mock, cleanup
}

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pwd123456"), bcrypt.DefaultCost)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at, password_hash`)).
		WithArgs("u@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "is_active", "created_at", "updated_at", "password_hash"}).
			AddRow(uid, "u@example.com", "U", "S", "user", true, time.Now(), time.Now(), string(hash)))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at)`)).
		WithArgs(uid, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := gin.New()
	r.POST("/login", func(c *gin.Context) {
		c.Set("validated_data", &models.LoginRequest{Email: "u@example.com", Password: "pwd123456"})
		h.Login(c)
	})

	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"email": "u@example.com", "password": "pwd123456"})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestAuthHandler_Login_MissingValidatedData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	r := gin.New()
	r.POST("/login", h.Login)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestAuthHandler_Login_InvalidCreds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at, password_hash`)).
		WithArgs("u@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "is_active", "created_at", "updated_at", "password_hash"}).
			AddRow(uid, "u@example.com", "U", "S", "user", true, time.Now(), time.Now(), string(hash)))

	r := gin.New()
	r.POST("/login", func(c *gin.Context) {
		c.Set("validated_data", &models.LoginRequest{Email: "u@example.com", Password: "wrong"})
		h.Login(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(`{"email":"u@example.com","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	uid := uuid.New()
	// verifyRefreshToken
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}).AddRow(uid, time.Now().Add(time.Hour)))
	// user fetch
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at`)).
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "is_active", "created_at", "updated_at"}).
			AddRow(uid, "u@example.com", "U", "S", "user", true, time.Now(), time.Now()))

	r := gin.New()
	r.POST("/refresh", func(c *gin.Context) {
		c.Set("validated_data", &models.RefreshTokenRequest{RefreshToken: "tok"})
		h.RefreshToken(c)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBufferString(`{"refresh_token":"tok"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestAuthHandler_Refresh_MissingValidatedData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	r := gin.New()
	r.POST("/refresh", h.RefreshToken)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestAuthHandler_Refresh_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	// verifyRefreshToken returns expired
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}).AddRow(uuid.New(), time.Now().Add(-time.Hour)))

	r := gin.New()
	r.POST("/refresh", func(c *gin.Context) {
		c.Set("validated_data", &models.RefreshTokenRequest{RefreshToken: "tok"})
		h.RefreshToken(c)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBufferString(`{"refresh_token":"tok"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuthHandler_CreateAPIKey_Fail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO api_keys (id, name, key_hash, permissions, is_active, created_at, expires_at)`)).
		WithArgs(sqlmock.AnyArg(), "key", sqlmock.AnyArg(), sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	r := gin.New()
	r.POST("/api-keys", h.CreateAPIKey)
	w := httptest.NewRecorder()
	exp := time.Now().Add(time.Hour)
	body, _ := json.Marshal(models.CreateAPIKeyRequest{Name: "key", Permissions: []string{"read"}, ExpiresAt: &exp})
	req, _ := http.NewRequest("POST", "/api-keys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestAuthHandler_CreateAPIKey_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO api_keys (id, name, key_hash, permissions, is_active, created_at, expires_at)`)).
		WithArgs(sqlmock.AnyArg(), "key", sqlmock.AnyArg(), sqlmock.AnyArg(), true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := gin.New()
	r.POST("/api-keys", h.CreateAPIKey)
	w := httptest.NewRecorder()
	exp := time.Now().Add(time.Hour)
	body, _ := json.Marshal(models.CreateAPIKeyRequest{Name: "key", Permissions: []string{"read"}, ExpiresAt: &exp})
	req, _ := http.NewRequest("POST", "/api-keys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", w.Code)
	}
}

func TestAuthHandler_CreateAPIKey_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, cleanup := newAuthHandlerForTest(t)
	defer cleanup()

	r := gin.New()
	r.POST("/api-keys", h.CreateAPIKey)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api-keys", bytes.NewBufferString(`{"name":`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestAuthHandler_GetProfile_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	h := &AuthHandler{logger: logger}
	r := gin.New()
	r.GET("/profile", h.GetProfile)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestAuthHandler_GetProfile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()
	// handler with nil service is fine for this path
	h := &AuthHandler{logger: logger}
	r := gin.New()
	r.GET("/profile", func(c *gin.Context) {
		c.Set("user_id", uuid.New().String())
		c.Set("user_email", "u@example.com")
		c.Set("user_role", models.UserRole("user"))
		h.GetProfile(c)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}
