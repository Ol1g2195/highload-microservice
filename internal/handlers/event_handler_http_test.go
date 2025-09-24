package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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

type stubKafkaEH struct{}

func (s *stubKafkaEH) SendEvent(ctx context.Context, event models.KafkaEvent) error { return nil }

type stubRedisEH struct{}

func (s *stubRedisEH) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (s *stubRedisEH) Get(ctx context.Context, key string) (string, error) { return "", sql.ErrNoRows }
func (s *stubRedisEH) Del(ctx context.Context, keys ...string) error       { return nil }

func newEventHandlerForTest(t *testing.T) (*EventHandler, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	svc := services.NewEventService(db, &stubRedisEH{}, &stubKafkaEH{}, logrus.New())
	h := NewEventHandler(svc, logrus.New())
	cleanup := func() { _ = db.Close() }
	return h, mock, cleanup
}

func TestEventHandler_CreateEvent_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newEventHandlerForTest(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	r := gin.New()
	r.POST("/events", h.CreateEvent)
	w := httptest.NewRecorder()
	body, _ := json.Marshal(models.CreateEventRequest{UserID: uuid.New(), Type: "t", Data: "{}"})
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", w.Code)
	}
}

func TestEventHandler_GetEvent_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newEventHandlerForTest(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at FROM events WHERE id = $1")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	r := gin.New()
	r.GET("/events/:id", h.GetEvent)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/events/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestEventHandler_ListEvents_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newEventHandlerForTest(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM events")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{"id", "user_id", "type", "data", "created_at"}).
		AddRow(uuid.New(), uuid.New(), "t", "{}", time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at ")).
		WillReturnRows(rows)

	r := gin.New()
	r.GET("/events", h.ListEvents)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/events?page=1&limit=10", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestEventHandler_ListEvents_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newEventHandlerForTest(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM events")).
		WillReturnError(sql.ErrConnDone)

	r := gin.New()
	r.GET("/events", h.ListEvents)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/events", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestEventHandler_ListEvents_PaginationBounds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newEventHandlerForTest(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM events")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at ")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "type", "data", "created_at"}))

	r := gin.New()
	r.GET("/events", h.ListEvents)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/events?page=-1&limit=1000", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestEventHandler_CreateEvent_Fail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mock, cleanup := newEventHandlerForTest(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	r := gin.New()
	r.POST("/events", h.CreateEvent)
	w := httptest.NewRecorder()
	body, _ := json.Marshal(models.CreateEventRequest{UserID: uuid.New(), Type: "t", Data: "{}"})
	req, _ := http.NewRequest("POST", "/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestEventHandler_GetEvent_BadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, cleanup := newEventHandlerForTest(t)
	defer cleanup()

	r := gin.New()
	r.GET("/events/:id", h.GetEvent)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/events/not-a-uuid", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
