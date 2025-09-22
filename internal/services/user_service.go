package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"highload-microservice/internal/kafka"
	"highload-microservice/internal/models"
	"highload-microservice/internal/redis"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	db            *sql.DB
	redisClient   *redis.Client
	kafkaProducer *kafka.Producer
	logger        *logrus.Logger
}

func NewUserService(db *sql.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, logger *logrus.Logger) *UserService {
	return &UserService{
		db:            db,
		redisClient:   redisClient,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	user := &models.User{
		ID:        uuid.New(),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO users (id, email, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Email, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Cache user data
	s.cacheUser(ctx, user)

	// Send event to Kafka
	event := models.KafkaEvent{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      "user_created",
		Data:      fmt.Sprintf(`{"email":"%s","first_name":"%s","last_name":"%s"}`, user.Email, user.FirstName, user.LastName),
		Timestamp: time.Now(),
	}

	if err := s.kafkaProducer.SendEvent(ctx, event); err != nil {
		s.logger.Errorf("Failed to send user creation event: %v", err)
	}

	s.logger.Infof("User created: %s", user.ID)
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:%s", id.String())
	if cached, err := s.redisClient.Get(ctx, cacheKey); err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			s.logger.Debugf("User %s retrieved from cache", id)
			return &user, nil
		}
	}

	// Get from database
	user := &models.User{}
	query := `SELECT id, email, first_name, last_name, created_at, updated_at FROM users WHERE id = $1`

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the result
	s.cacheUser(ctx, user)

	s.logger.Debugf("User %s retrieved from database", id)
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users 
		SET email = $1, first_name = $2, last_name = $3, updated_at = $4
		WHERE id = $5
	`

	_, err = s.db.ExecContext(ctx, query, user.Email, user.FirstName, user.LastName, user.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Update cache
	s.cacheUser(ctx, user)

	// Send event to Kafka
	event := models.KafkaEvent{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      "user_updated",
		Data:      fmt.Sprintf(`{"email":"%s","first_name":"%s","last_name":"%s"}`, user.Email, user.FirstName, user.LastName),
		Timestamp: time.Now(),
	}

	if err := s.kafkaProducer.SendEvent(ctx, event); err != nil {
		s.logger.Errorf("Failed to send user update event: %v", err)
	}

	s.logger.Infof("User updated: %s", id)
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("user:%s", id.String())
	_ = s.redisClient.Del(ctx, cacheKey) // Ignore cache deletion errors

	// Send event to Kafka
	event := models.KafkaEvent{
		ID:        uuid.New(),
		UserID:    id,
		Type:      "user_deleted",
		Data:      `{}`,
		Timestamp: time.Now(),
	}

	if err := s.kafkaProducer.SendEvent(ctx, event); err != nil {
		s.logger.Errorf("Failed to send user deletion event: %v", err)
	}

	s.logger.Infof("User deleted: %s", id)
	return nil
}

func (s *UserService) ListUsers(ctx context.Context, page, limit int) (*models.UserListResponse, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	err := s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	query := `
		SELECT id, email, first_name, last_name, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return &models.UserListResponse{
		Users: users,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *UserService) cacheUser(ctx context.Context, user *models.User) {
	cacheKey := fmt.Sprintf("user:%s", user.ID.String())
	userData, err := json.Marshal(user)
	if err != nil {
		s.logger.Errorf("Failed to marshal user for cache: %v", err)
		return
	}

	if err := s.redisClient.Set(ctx, cacheKey, string(userData), 1*time.Hour); err != nil {
		s.logger.Errorf("Failed to cache user: %v", err)
	}
}


