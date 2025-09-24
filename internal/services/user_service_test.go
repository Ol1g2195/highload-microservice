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

type stubRedis struct{}

func (s *stubRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (s *stubRedis) Get(ctx context.Context, key string) (string, error)  { return "", sql.ErrNoRows }
func (s *stubRedis) Del(ctx context.Context, keys ...string) error        { return nil }
func (s *stubRedis) Exists(ctx context.Context, key string) (bool, error) { return false, nil }
func (s *stubRedis) Ping(ctx context.Context) error                       { return nil }
func (s *stubRedis) Close() error                                         { return nil }

type stubProducer struct{}

func (s *stubProducer) SendEvent(ctx context.Context, _ models.KafkaEvent) error { return nil }
func (s *stubProducer) Close() error                                             { return nil }

// compile-time checks that stubs satisfy minimal interfaces used in service
var _ = (&stubRedis{}).Ping
var _ = (&stubProducer{}).Close

func TestUserService_CreateAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	logger := logrus.New()
	svc := &UserService{db: db, redisClient: &stubRedis{}, kafkaProducer: &stubProducer{}, logger: logger}

	// Insert expectation
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(sqlmock.AnyArg(), "u@example.com", "John", "Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create
	user, err := svc.CreateUser(context.Background(), models.CreateUserRequest{
		Email:     "u@example.com",
		FirstName: "John",
		LastName:  "Doe",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Query expectation for GetUser
	rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
		AddRow(user.ID, user.Email, user.FirstName, user.LastName, time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(user.ID).WillReturnRows(rows)

	got, err := svc.GetUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Email != "u@example.com" {
		t.Fatalf("unexpected email: %s", got.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestUserService_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := &UserService{db: db, logger: logrus.New()}

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = svc.DeleteUser(context.Background(), uuid.New())
	if err == nil {
		t.Fatalf("expected not found error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestUserService_Delete_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := &UserService{db: db, logger: logrus.New()}

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected failed")))

	if err := svc.DeleteUser(context.Background(), uuid.New()); err == nil {
		t.Fatalf("expected rows affected error")
	}
}

func TestUserService_ListUsers_Errors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := &UserService{db: db, logger: logrus.New()}

	// count error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
		WillReturnError(fmt.Errorf("count failed"))
	if _, err := svc.ListUsers(context.Background(), 1, 10); err == nil {
		t.Fatalf("expected count error")
	}

	// query error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at ")).
		WillReturnError(fmt.Errorf("list failed"))
	if _, err := svc.ListUsers(context.Background(), 1, 10); err == nil {
		t.Fatalf("expected list query error")
	}

	// scan error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at ")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
			AddRow("not-uuid", "e@x", "f", "l", time.Now(), time.Now()))
	if _, err := svc.ListUsers(context.Background(), 1, 10); err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestUserService_ListUsers_SuccessMultiple(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := &UserService{db: db, logger: logrus.New()}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
		AddRow(uuid.New(), "a@example.com", "A", "A", time.Now(), time.Now()).
		AddRow(uuid.New(), "b@example.com", "B", "B", time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at ")).
		WillReturnRows(rows)

	out, err := svc.ListUsers(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if out.Total != 2 || len(out.Users) != 2 {
		t.Fatalf("unexpected list result")
	}
}

func TestUserService_UpdateUser_NotFoundAndDBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := &UserService{db: db, redisClient: &stubRedis{}, kafkaProducer: &stubProducer{}, logger: logrus.New()}
	id := uuid.New()

	// GetUser not found
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(id).WillReturnError(sql.ErrNoRows)
	if _, err := svc.UpdateUser(context.Background(), id, models.UpdateUserRequest{}); err == nil {
		t.Fatalf("expected not found from GetUser")
	}

	// Successful GetUser, then update DB error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
			AddRow(id, "u@example.com", "J", "D", time.Now(), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users ")).
		WithArgs("u@example.com", "J", "D", sqlmock.AnyArg(), id).
		WillReturnError(fmt.Errorf("update failed"))
	if _, err := svc.UpdateUser(context.Background(), id, models.UpdateUserRequest{}); err == nil {
		t.Fatalf("expected update failed")
	}
}

type stubRedisWithValue struct{ val string }

func (s *stubRedisWithValue) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (s *stubRedisWithValue) Get(ctx context.Context, key string) (string, error)  { return s.val, nil }
func (s *stubRedisWithValue) Del(ctx context.Context, keys ...string) error        { return nil }
func (s *stubRedisWithValue) Exists(ctx context.Context, key string) (bool, error) { return true, nil }
func (s *stubRedisWithValue) Ping(ctx context.Context) error                       { return nil }
func (s *stubRedisWithValue) Close() error                                         { return nil }

func TestUserService_GetUser_CacheHit(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	u := models.User{ID: uuid.New(), Email: "c@example.com", FirstName: "C", LastName: "H", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	buf, _ := json.Marshal(u)
	svc := &UserService{db: db, redisClient: &stubRedisWithValue{val: string(buf)}, kafkaProducer: &stubProducer{}, logger: logrus.New()}

	got, err := svc.GetUser(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("cache get: %v", err)
	}
	if got.Email != u.Email {
		t.Fatalf("email mismatch: %s", got.Email)
	}
}

type stubRedisCorrupt struct{}

func (s *stubRedisCorrupt) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (s *stubRedisCorrupt) Get(ctx context.Context, key string) (string, error) {
	return "{not-json}", nil
}
func (s *stubRedisCorrupt) Del(ctx context.Context, keys ...string) error        { return nil }
func (s *stubRedisCorrupt) Exists(ctx context.Context, key string) (bool, error) { return false, nil }
func (s *stubRedisCorrupt) Ping(ctx context.Context) error                       { return nil }
func (s *stubRedisCorrupt) Close() error                                         { return nil }

func TestUserService_GetUser_DBErrorAndCorruptCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := &UserService{db: db, redisClient: &stubRedisCorrupt{}, kafkaProducer: &stubProducer{}, logger: logrus.New()}
	id := uuid.New()

	// Non-ErrNoRows DB error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(id).WillReturnError(fmt.Errorf("db failure"))
	if _, err := svc.GetUser(context.Background(), id); err == nil {
		t.Fatalf("expected db failure")
	}

	// Success after corrupt cache (fallback to DB)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
			AddRow(id, "ok@example.com", "F", "L", time.Now(), time.Now()))
	u, err := svc.GetUser(context.Background(), id)
	if err != nil {
		t.Fatalf("get after corrupt cache: %v", err)
	}
	if u.Email != "ok@example.com" {
		t.Fatalf("unexpected email: %s", u.Email)
	}
}

type stubProducerErr struct{}

func (s *stubProducerErr) SendEvent(ctx context.Context, _ models.KafkaEvent) error {
	return fmt.Errorf("kafka down")
}
func (s *stubProducerErr) Close() error { return nil }

type stubRedisErr struct{}

func (s *stubRedisErr) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return fmt.Errorf("set failed")
}
func (s *stubRedisErr) Get(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("get failed")
}
func (s *stubRedisErr) Del(ctx context.Context, keys ...string) error {
	return fmt.Errorf("del failed")
}
func (s *stubRedisErr) Exists(ctx context.Context, key string) (bool, error) {
	return false, fmt.Errorf("exists failed")
}
func (s *stubRedisErr) Ping(ctx context.Context) error { return nil }
func (s *stubRedisErr) Close() error                   { return nil }

func TestUserService_Create_Update_Delete_WithKafkaRedisErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	logger := logrus.New()
	svc := &UserService{db: db, redisClient: &stubRedisErr{}, kafkaProducer: &stubProducerErr{}, logger: logger}

	// CreateUser still succeeds even if cache/kafka fail
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(sqlmock.AnyArg(), "e@x", "F", "L", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	u, err := svc.CreateUser(context.Background(), models.CreateUserRequest{Email: "e@x", FirstName: "F", LastName: "L"})
	if err != nil {
		t.Fatalf("create err: %v", err)
	}

	// UpdateUser: GetUser from DB then UPDATE; cache/kafka errors are logged only
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs(u.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "created_at", "updated_at"}).
			AddRow(u.ID, u.Email, u.FirstName, u.LastName, time.Now(), time.Now()))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users ")).
		WithArgs(u.Email, u.FirstName, u.LastName, sqlmock.AnyArg(), u.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := svc.UpdateUser(context.Background(), u.ID, models.UpdateUserRequest{}); err != nil {
		t.Fatalf("update err: %v", err)
	}

	// DeleteUser: DELETE returns 1 row; redis Del fails but method returns nil
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(u.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := svc.DeleteUser(context.Background(), u.ID); err != nil {
		t.Fatalf("delete err: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
