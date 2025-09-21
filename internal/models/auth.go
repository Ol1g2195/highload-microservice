package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UserRole represents user roles in the system
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleUser     UserRole = "user"
	RoleReadOnly UserRole = "readonly"
)

// AuthUser represents an authenticated user
type AuthUser struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Role      UserRole  `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" validate:"required,email,email_domain,no_sql_injection,no_xss"`
	Password string `json:"password" binding:"required,min=8" validate:"required,min=8,max=128,no_sql_injection,no_xss"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         AuthUser `json:"user"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" validate:"required,min=32,max=128,safe_string,no_sql_injection,no_xss"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	ExpiresAt int64     `json:"exp"`
	IssuedAt  int64     `json:"iat"`
	Issuer    string    `json:"iss"`
}

// GetAudience implements jwt.Claims
func (c JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}

// GetExpirationTime implements jwt.Claims
func (c JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.ExpiresAt, 0)), nil
}

// GetIssuedAt implements jwt.Claims
func (c JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

// GetIssuer implements jwt.Claims
func (c JWTClaims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

// GetNotBefore implements jwt.Claims
func (c JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetSubject implements jwt.Claims
func (c JWTClaims) GetSubject() (string, error) {
	return c.UserID.String(), nil
}

// APIKey represents API key for service-to-service authentication
type APIKey struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	KeyHash     string     `json:"-" db:"key_hash"`
	Permissions []string   `json:"permissions" db:"permissions"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
}

// CreateAPIKeyRequest represents API key creation request
type CreateAPIKeyRequest struct {
	Name        string     `json:"name" binding:"required,min=3,max=50" validate:"required,min=3,max=50,safe_string,no_sql_injection,no_xss"`
	Permissions []string   `json:"permissions" binding:"required" validate:"required,min=1,dive,required,safe_string,no_sql_injection,no_xss"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// CreateAPIKeyResponse represents API key creation response
type CreateAPIKeyResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	APIKey    string     `json:"api_key"` // Only shown once during creation
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}
