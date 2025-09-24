package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"highload-microservice/internal/models"
	"highload-microservice/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type stubRedis struct{}

func (s *stubRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (s *stubRedis) Get(ctx context.Context, key string) (string, error) { return "", sql.ErrNoRows }
func (s *stubRedis) Del(ctx context.Context, keys ...string) error       { return nil }

type stubKafka struct{}

func (s *stubKafka) SendEvent(ctx context.Context, event models.KafkaEvent) error { return nil }

func newUserHandler(t *testing.T) (*UserHandler, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	logger := logrus.New()
	svc := services.NewUserService(db, &stubRedis{}, &stubKafka{}, logger)
	h := NewUserHandler(svc, logger)
	cleanup := func() { db.Close() }
	return h, mock, cleanup
}

func TestUserHandler_CreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(sqlmock.AnyArg(), "u@example.com", "John", "Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := gin.New()
	r.POST("/users", func(c *gin.Context) {
		c.Set("validated_data", &models.CreateUserRequest{Email: "u@example.com", FirstName: "John", LastName: "Doe"})
		h.CreateUser(c)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	r := gin.New()
	r.GET("/users/:id", h.GetUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUserHandler_GetUser_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, cleanup := newUserHandler(t)
	defer cleanup()

	r := gin.New()
	r.GET("/users/:id", h.GetUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/not-uuid", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
		AddRow(uuid.New(), "u@example.com", "J", "D", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at ")).
		WillReturnRows(rows)

	r := gin.New()
	r.GET("/users", h.ListUsers)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}

	var out []models.User
	_ = json.Unmarshal(w.Body.Bytes(), &out)
}

func TestUserHandler_ListUsers_PaginationBounds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at ")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}))

	r := gin.New()
	r.GET("/users", h.ListUsers)
	w := httptest.NewRecorder()
	// Некорректные page/limit должны замениться на значения по умолчанию
	req, _ := http.NewRequest("GET", "/users?page=-10&limit=1000", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserHandler_UpdateUser_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, cleanup := newUserHandler(t)
	defer cleanup()

	r := gin.New()
	r.PUT("/users/:id", h.UpdateUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/not-uuid", bytes.NewBufferString(`{"email":"a@b"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_UpdateUser_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, cleanup := newUserHandler(t)
	defer cleanup()

	id := uuid.New()
	r := gin.New()
	r.PUT("/users/:id", h.UpdateUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/"+id.String(), bytes.NewBufferString(`{"email":`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_DeleteUser_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, cleanup := newUserHandler(t)
	defer cleanup()

	r := gin.New()
	r.DELETE("/users/:id", h.DeleteUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/not-uuid", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_CreateUser_Conflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(sqlmock.AnyArg(), "u@example.com", "John", "Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("duplicate key value violates unique constraint (SQLSTATE 23505)"))

	r := gin.New()
	r.POST("/users", func(c *gin.Context) {
		c.Set("validated_data", &models.CreateUserRequest{Email: "u@example.com", FirstName: "John", LastName: "Doe"})
		h.CreateUser(c)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestUserHandler_CreateUser_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(sqlmock.AnyArg(), "u@example.com", "John", "Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("db down"))

	r := gin.New()
	r.POST("/users", func(c *gin.Context) {
		c.Set("validated_data", &models.CreateUserRequest{Email: "u@example.com", FirstName: "John", LastName: "Doe"})
		h.CreateUser(c)
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUserHandler_ListUsers_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
		WillReturnError(fmt.Errorf("count failed"))

	r := gin.New()
	r.GET("/users", h.ListUsers)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUserHandler_UpdateUser_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	id := uuid.New()
	// GetUser SELECT
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
			AddRow(id, "u@example.com", "J", "D", time.Now(), time.Now()))
	// UPDATE
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users ")).
		WithArgs("new@example.com", "J", "D", sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := gin.New()
	r.PUT("/users/:id", h.UpdateUser)
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"email":"new@example.com"}`)
	req, _ := http.NewRequest("PUT", "/users/"+id.String(), body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUserHandler_UpdateUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	id := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(id).WillReturnError(sql.ErrNoRows)

	r := gin.New()
	r.PUT("/users/:id", h.UpdateUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/"+id.String(), bytes.NewBufferString(`{"email":"a@b"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUserHandler_UpdateUser_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	id := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
			AddRow(id, "u@example.com", "J", "D", time.Now(), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users ")).
		WithArgs("u@example.com", "J", "D", sqlmock.AnyArg(), id).
		WillReturnError(fmt.Errorf("db failed"))

	r := gin.New()
	r.PUT("/users/:id", h.UpdateUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/users/"+id.String(), bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUserHandler_DeleteUser_NoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	id := uuid.New()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := gin.New()
	r.DELETE("/users/:id", h.DeleteUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/"+id.String(), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestUserHandler_DeleteUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	id := uuid.New()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 0))

	r := gin.New()
	r.DELETE("/users/:id", h.DeleteUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/"+id.String(), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUserHandler_DeleteUser_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newUserHandler(t)
	defer cleanup()

	id := uuid.New()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(id).
		WillReturnError(fmt.Errorf("db failed"))

	r := gin.New()
	r.DELETE("/users/:id", h.DeleteUser)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/users/"+id.String(), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
