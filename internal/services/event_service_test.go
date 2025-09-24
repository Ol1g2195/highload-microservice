package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"highload-microservice/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type stubRedisGetSet struct{}

func (s *stubRedisGetSet) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (s *stubRedisGetSet) Get(ctx context.Context, key string) (string, error) {
	return "", sql.ErrNoRows
}
func (s *stubRedisGetSet) Del(ctx context.Context, keys ...string) error { return nil }

type stubKafka struct{}

func (s *stubKafka) SendEvent(ctx context.Context, event models.KafkaEvent) error { return nil }

// redisHit returns cached payload for Get
type redisHitWithPayload struct{ payload string }

func (r *redisHitWithPayload) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (r *redisHitWithPayload) Get(ctx context.Context, key string) (string, error) {
	return r.payload, nil
}
func (r *redisHitWithPayload) Del(ctx context.Context, keys ...string) error { return nil }

func TestEventService_CreateAndList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := NewEventService(db, &stubRedisGetSet{}, &stubKafka{}, logrus.New())

	// Create
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = svc.CreateEvent(context.Background(), models.CreateEventRequest{UserID: uuid.New(), Type: "created", Data: "{}"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// List
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM events")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{"id", "user_id", "type", "data", "created_at"}).
		AddRow(uuid.New(), uuid.New(), "created", "{}", time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at ")).
		WillReturnRows(rows)

	list, err := svc.ListEvents(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if list.Total != 1 || len(list.Events) != 1 {
		t.Fatalf("unexpected list result")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

type kafkaErr struct{}

func (k *kafkaErr) SendEvent(ctx context.Context, event models.KafkaEvent) error {
	return fmt.Errorf("kafka down")
}

type redisErr struct{}

func (r *redisErr) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return fmt.Errorf("set err")
}
func (r *redisErr) Get(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("get err")
}
func (r *redisErr) Del(ctx context.Context, keys ...string) error { return fmt.Errorf("del err") }

func TestEventService_Create_WithKafkaRedisErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := NewEventService(db, &redisErr{}, &kafkaErr{}, logrus.New())

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO events")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := svc.CreateEvent(context.Background(), models.CreateEventRequest{UserID: uuid.New(), Type: "t", Data: "{}"}); err != nil {
		t.Fatalf("create should succeed despite kafka/redis errors: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// cache miss -> DB success -> cache set error (already covered by redisErr.Set), ensure method still succeeds
func TestEventService_GetEvent_CacheMissThenDB_SetsCacheError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// redisErr.Get returns error -> cache miss; Set will also error
	svc := NewEventService(db, &redisErr{}, &stubKafka{}, logrus.New())

	id := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at FROM events WHERE id = $1")).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "type", "data", "created_at"}).
			AddRow(id, uuid.New(), "t", "{}", time.Now()))

	if ev, err := svc.GetEvent(context.Background(), id); err != nil || ev == nil {
		t.Fatalf("expected success from DB with cache errors, err=%v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestEventService_GetEvent_CacheHit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// cached event
	e := models.Event{ID: uuid.New(), UserID: uuid.New(), Type: "t", Data: "{}", CreatedAt: time.Now()}
	payload, _ := json.Marshal(e)
	rh := &redisHitWithPayload{payload: string(payload)}

	svc := NewEventService(db, rh, &stubKafka{}, logrus.New())

	// No DB expectations; should return from cache directly
	got, err := svc.GetEvent(context.Background(), e.ID)
	if err != nil {
		t.Fatalf("cache get: %v", err)
	}
	if got == nil || got.ID != e.ID {
		t.Fatalf("unexpected event from cache")
	}

	if err := mock.ExpectationsWereMet(); err != nil { /* none expected */
	}
}

func TestEventService_ListEvents_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	svc := NewEventService(db, &stubRedisGetSet{}, &stubKafka{}, logrus.New())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM events")).
		WillReturnError(sql.ErrConnDone)

	if _, err := svc.ListEvents(context.Background(), 1, 10); err == nil {
		t.Fatalf("expected error on count")
	}
}

func TestEventService_ListEvents_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	svc := NewEventService(db, &stubRedisGetSet{}, &stubKafka{}, logrus.New())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM events")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at ")).
		WillReturnError(sql.ErrConnDone)

	if _, err := svc.ListEvents(context.Background(), 1, 10); err == nil {
		t.Fatalf("expected error on list query")
	}
}

func TestEventService_ListEvents_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	svc := NewEventService(db, &stubRedisGetSet{}, &stubKafka{}, logrus.New())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM events")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// wrong column types to trigger scan error
	rows := sqlmock.NewRows([]string{"id", "user_id", "type", "data", "created_at"}).
		AddRow("not-uuid", "not-uuid", 123, 456, "not-time")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at ")).
		WillReturnRows(rows)

	if _, err := svc.ListEvents(context.Background(), 1, 10); err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestEventService_GetEvent_DBErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	svc := NewEventService(db, &stubRedisGetSet{}, &stubKafka{}, logrus.New())

	id := uuid.New()
	// not found
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at FROM events WHERE id = $1")).
		WithArgs(id).WillReturnError(sql.ErrNoRows)
	if _, err := svc.GetEvent(context.Background(), id); err == nil {
		t.Fatalf("expected not found")
	}

	// other DB error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, type, data, created_at FROM events WHERE id = $1")).
		WithArgs(id).WillReturnError(sql.ErrConnDone)
	if _, err := svc.GetEvent(context.Background(), id); err == nil {
		t.Fatalf("expected db error")
	}
}
