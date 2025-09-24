package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"highload-microservice/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db     *sql.DB
	logger *logrus.Logger
	config AuthConfig
}

type AuthConfig struct {
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration
	APIKeyLength      int
}

func NewAuthService(db *sql.DB, logger *logrus.Logger, config AuthConfig) *AuthService {
	return &AuthService{
		db:     db,
		logger: logger,
		config: config,
	}
}

// AuthenticateUser authenticates user with email and password
func (s *AuthService) AuthenticateUser(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by email
	var user models.AuthUser
	var passwordHash string

	query := `SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at, password_hash 
			  FROM auth_users WHERE email = $1 AND is_active = true`

	err := s.db.QueryRowContext(ctx, query, req.Email).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &passwordHash,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warnf("Authentication failed for email: %s - user not found", req.Email)
			return nil, fmt.Errorf("invalid credentials")
		}
		s.logger.Errorf("Database error during authentication: %v", err)
		return nil, fmt.Errorf("authentication failed")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		s.logger.Warnf("Authentication failed for email: %s - invalid password", req.Email)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		s.logger.Errorf("Failed to generate access token: %v", err)
		return nil, fmt.Errorf("token generation failed")
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		s.logger.Errorf("Failed to generate refresh token: %v", err)
		return nil, fmt.Errorf("token generation failed")
	}

	// Store refresh token in database
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken); err != nil {
		s.logger.Errorf("Failed to store refresh token: %v", err)
		return nil, fmt.Errorf("token storage failed")
	}

	s.logger.Infof("User authenticated successfully: %s", user.Email)

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.JWTExpiration.Seconds()),
		User:         user,
	}, nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (*models.LoginResponse, error) {
	// Verify refresh token
	userID, err := s.verifyRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Warnf("Invalid refresh token: %v", err)
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user
	var user models.AuthUser
	query := `SELECT id, email, first_name, last_name, role, is_active, created_at, updated_at 
			  FROM auth_users WHERE id = $1 AND is_active = true`

	err = s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		s.logger.Errorf("Failed to get user for refresh: %v", err)
		return nil, fmt.Errorf("user not found")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		s.logger.Errorf("Failed to generate new access token: %v", err)
		return nil, fmt.Errorf("token generation failed")
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken, // Keep the same refresh token
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.JWTExpiration.Seconds()),
		User:         user,
	}, nil
}

// ValidateToken validates JWT token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Convert map claims to JWTClaims
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid user_id in token")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id format: %v", err)
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid email in token")
		}

		roleStr, ok := claims["role"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid role in token")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid exp in token")
		}

		iat, ok := claims["iat"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid iat in token")
		}

		iss, ok := claims["iss"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid iss in token")
		}

		// Check if token is expired
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}

		return &models.JWTClaims{
			UserID:    userID,
			Email:     email,
			Role:      models.UserRole(roleStr),
			ExpiresAt: int64(exp),
			IssuedAt:  int64(iat),
			Issuer:    iss,
		}, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// CreateAPIKey creates a new API key
func (s *AuthService) CreateAPIKey(ctx context.Context, req models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	// Generate API key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		s.logger.Errorf("Failed to generate API key: %v", err)
		return nil, fmt.Errorf("failed to generate API key")
	}

	// Hash the API key for storage
	keyHash := s.hashAPIKey(apiKey)

	// Create API key record
	apiKeyID := uuid.New()
	query := `INSERT INTO api_keys (id, name, key_hash, permissions, is_active, created_at, expires_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = s.db.ExecContext(ctx, query, apiKeyID, req.Name, keyHash, pq.Array(req.Permissions), true, time.Now(), req.ExpiresAt)
	if err != nil {
		s.logger.Errorf("Failed to create API key: %v", err)
		return nil, fmt.Errorf("failed to create API key")
	}

	s.logger.Infof("API key created: %s", req.Name)

	return &models.CreateAPIKeyResponse{
		ID:        apiKeyID,
		Name:      req.Name,
		APIKey:    apiKey, // Only returned once
		ExpiresAt: req.ExpiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// ValidateAPIKey validates API key and returns permissions
func (s *AuthService) ValidateAPIKey(ctx context.Context, apiKey string) ([]string, error) {
	keyHash := s.hashAPIKey(apiKey)

	var permissions pq.StringArray
	var isActive bool
	var expiresAt *time.Time

	query := `SELECT permissions, is_active, expires_at FROM api_keys WHERE key_hash = $1`
	err := s.db.QueryRowContext(ctx, query, keyHash).Scan(&permissions, &isActive, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		s.logger.Errorf("Database error during API key validation: %v", err)
		return nil, fmt.Errorf("API key validation failed")
	}

	if !isActive {
		return nil, fmt.Errorf("API key is inactive")
	}

	if expiresAt != nil && time.Now().After(*expiresAt) {
		return nil, fmt.Errorf("API key expired")
	}

	return []string(permissions), nil
}

// Helper methods

func (s *AuthService) generateAccessToken(user models.AuthUser) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    string(user.Role),
		"exp":     now.Add(s.config.JWTExpiration).Unix(),
		"iat":     now.Unix(),
		"iss":     "highload-microservice",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *AuthService) generateRefreshToken(userID uuid.UUID) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AuthService) storeRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at) 
			  VALUES ($1, $2, $3, $4)`

	tokenHash := s.hashAPIKey(token) // Reuse hash function
	expiresAt := time.Now().Add(s.config.RefreshExpiration)

	_, err := s.db.ExecContext(ctx, query, userID, tokenHash, expiresAt, time.Now())
	return err
}

func (s *AuthService) verifyRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	tokenHash := s.hashAPIKey(token)

	var userID uuid.UUID
	var expiresAt time.Time

	query := `SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`
	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(&userID, &expiresAt)

	if err != nil {
		return uuid.Nil, err
	}

	if time.Now().After(expiresAt) {
		return uuid.Nil, fmt.Errorf("refresh token expired")
	}

	return userID, nil
}

func (s *AuthService) generateAPIKey() (string, error) {
	bytes := make([]byte, s.config.APIKeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "hl_" + hex.EncodeToString(bytes), nil
}

func (s *AuthService) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
