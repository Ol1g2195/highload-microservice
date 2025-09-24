package services

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"highload-microservice/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func newAuthServiceMock(t *testing.T) (*AuthService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	cfg := AuthConfig{JWTSecret: "secret", JWTExpiration: time.Hour, RefreshExpiration: 24 * time.Hour, APIKeyLength: 4}
	svc := NewAuthService(db, logrus.New(), cfg)
	cleanup := func() { db.Close() }
	return svc, mock, cleanup
}

func TestAuthenticateUser_Success(t *testing.T) {
	svc, mock, cleanup := newAuthServiceMock(t)
	defer cleanup()

	uid := uuid.New()
	// bcrypt password: hash of "admin123456"
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123456"), bcrypt.DefaultCost)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at, password_hash 
              FROM auth_users WHERE email = $1 AND is_active = true`)).
		WithArgs("admin@local").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "is_active", "created_at", "updated_at", "password_hash"}).
			AddRow(uid, "admin@local", "Admin", "User", "admin", true, time.Now(), time.Now(), string(hash)))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at) 
              VALUES ($1, $2, $3, $4)`)).
		WithArgs(uid, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := svc.AuthenticateUser(context.Background(), models.LoginRequest{Email: "admin@local", Password: "admin123456"})
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("tokens not returned")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	svc, mock, cleanup := newAuthServiceMock(t)
	defer cleanup()

	uid := uuid.New()
	// prepare stored refresh token
	tok := "abcdef"
	// Expect verifyRefreshToken query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`)).
		WithArgs(svc.hashAPIKey(tok)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}).AddRow(uid, time.Now().Add(time.Hour)))

	// Expect user fetch
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at 
              FROM auth_users WHERE id = $1 AND is_active = true`)).
		WithArgs(uid).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "is_active", "created_at", "updated_at"}).
			AddRow(uid, "admin@local", "Admin", "User", "admin", true, time.Now(), time.Now()))

	resp, err := svc.RefreshToken(context.Background(), models.RefreshTokenRequest{RefreshToken: tok})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatalf("no new access token")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestAuthenticateUser_InvalidPassword(t *testing.T) {
	svc, mock, cleanup := newAuthServiceMock(t)
	defer cleanup()

	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at, password_hash 
              FROM auth_users WHERE email = $1 AND is_active = true`)).
		WithArgs("user@local").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "is_active", "created_at", "updated_at", "password_hash"}).
			AddRow(uid, "user@local", "U", "S", "user", true, time.Now(), time.Now(), string(hash)))

	_, err := svc.AuthenticateUser(context.Background(), models.LoginRequest{Email: "user@local", Password: "wrong"})
	if err == nil {
		t.Fatalf("expected invalid credentials")
	}
}

func TestAuthenticateUser_DBError(t *testing.T) {
	svc, mock, cleanup := newAuthServiceMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at, password_hash 
             FROM auth_users WHERE email = $1 AND is_active = true`)).
		WithArgs("u@example.com").
		WillReturnError(fmt.Errorf("db down"))

	_, err := svc.AuthenticateUser(context.Background(), models.LoginRequest{Email: "u@example.com", Password: "x"})
	if err == nil || !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("expected authentication failed, got %v", err)
	}
}

func TestValidateAPIKey_NotFound(t *testing.T) {
	svc, mock, cleanup := newAuthServiceMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT permissions, is_active, expires_at FROM api_keys WHERE key_hash = $1`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.ValidateAPIKey(context.Background(), "hl_abc")
	if err == nil || !strings.Contains(err.Error(), "invalid API key") {
		t.Fatalf("expected invalid API key, got %v", err)
	}
}

func TestValidateAPIKey_DBError(t *testing.T) {
	svc, mock, cleanup := newAuthServiceMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT permissions, is_active, expires_at FROM api_keys WHERE key_hash = $1`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("db err"))

	_, err := svc.ValidateAPIKey(context.Background(), "hl_abc")
	if err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failed, got %v", err)
	}
}

func TestValidateAPIKey_Expired(t *testing.T) {
	svc, mock, cleanup := newAuthServiceMock(t)
	defer cleanup()

	past := time.Now().Add(-time.Hour)
	rows := sqlmock.NewRows([]string{"permissions", "is_active", "expires_at"}).
		AddRow(pq.Array([]string{"read"}), true, past)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT permissions, is_active, expires_at FROM api_keys WHERE key_hash = $1`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	_, err := svc.ValidateAPIKey(context.Background(), "hl_abc")
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired, got %v", err)
	}
}

func TestValidateToken_SuccessAndInvalid(t *testing.T) {
	svc, _, cleanup := newAuthServiceMock(t)
	defer cleanup()

	user := models.AuthUser{ID: uuid.New(), Email: "u@l", Role: "user"}
	tok, err := svc.generateAccessToken(user)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := svc.ValidateToken(tok); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if _, err := svc.ValidateToken("not-a-token"); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}

func TestValidateToken_MissingAndBadClaims(t *testing.T) {
	svc, _, cleanup := newAuthServiceMock(t)
	defer cleanup()

	makeTok := func(claims jwt.MapClaims) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		s, _ := tok.SignedString([]byte(svc.config.JWTSecret))
		return s
	}

	now := time.Now()
	user := models.AuthUser{ID: uuid.New(), Email: "u@l", Role: "user"}
	base := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    string(user.Role),
		"exp":     now.Add(time.Hour).Unix(),
		"iat":     now.Unix(),
		"iss":     "highload-microservice",
	}

	// missing user_id
	c1 := jwt.MapClaims{}
	for k, v := range base {
		c1[k] = v
	}
	delete(c1, "user_id")
	if _, err := svc.ValidateToken(makeTok(c1)); err == nil {
		t.Fatalf("expected error for missing user_id")
	}

	// bad user_id format
	c2 := jwt.MapClaims{}
	for k, v := range base {
		c2[k] = v
	}
	c2["user_id"] = "not-uuid"
	if _, err := svc.ValidateToken(makeTok(c2)); err == nil {
		t.Fatalf("expected error for bad user_id format")
	}

	// missing email
	c3 := jwt.MapClaims{}
	for k, v := range base {
		c3[k] = v
	}
	delete(c3, "email")
	if _, err := svc.ValidateToken(makeTok(c3)); err == nil {
		t.Fatalf("expected error for missing email")
	}

	// missing role
	c4 := jwt.MapClaims{}
	for k, v := range base {
		c4[k] = v
	}
	delete(c4, "role")
	if _, err := svc.ValidateToken(makeTok(c4)); err == nil {
		t.Fatalf("expected error for missing role")
	}

	// missing exp
	c5 := jwt.MapClaims{}
	for k, v := range base {
		c5[k] = v
	}
	delete(c5, "exp")
	if _, err := svc.ValidateToken(makeTok(c5)); err == nil {
		t.Fatalf("expected error for missing exp")
	}

	// missing iat
	c6 := jwt.MapClaims{}
	for k, v := range base {
		c6[k] = v
	}
	delete(c6, "iat")
	if _, err := svc.ValidateToken(makeTok(c6)); err == nil {
		t.Fatalf("expected error for missing iat")
	}

	// missing iss
	c7 := jwt.MapClaims{}
	for k, v := range base {
		c7[k] = v
	}
	delete(c7, "iss")
	if _, err := svc.ValidateToken(makeTok(c7)); err == nil {
		t.Fatalf("expected error for missing iss")
	}
}

func TestRefreshToken_Expired(t *testing.T) {
	svc, mock, cleanup := newAuthServiceMock(t)
	defer cleanup()

	tok := "expired"
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`)).
		WithArgs(svc.hashAPIKey(tok)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}).AddRow(uuid.New(), time.Now().Add(-time.Hour)))

	_, err := svc.RefreshToken(context.Background(), models.RefreshTokenRequest{RefreshToken: tok})
	if err == nil {
		t.Fatalf("expected expired refresh token error")
	}
}
