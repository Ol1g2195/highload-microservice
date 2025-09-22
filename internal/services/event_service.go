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

type EventService struct {
	db            *sql.DB
	redisClient   *redis.Client
	kafkaProducer *kafka.Producer
	logger        *logrus.Logger
}

func NewEventService(db *sql.DB, redisClient *redis.Client, kafkaProducer *kafka.Producer, logger *logrus.Logger) *EventService {
	return &EventService{
		db:            db,
		redisClient:   redisClient,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}
}

func (s *EventService) CreateEvent(ctx context.Context, req models.CreateEventRequest) (*models.Event, error) {
	event := &models.Event{
		ID:        uuid.New(),
		UserID:    req.UserID,
		Type:      req.Type,
		Data:      req.Data,
		CreatedAt: time.Now(),
	}

	query := `
		INSERT INTO events (id, user_id, type, data, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.ExecContext(ctx, query, event.ID, event.UserID, event.Type, event.Data, event.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Send event to Kafka
	kafkaEvent := models.KafkaEvent{
		ID:        event.ID,
		UserID:    event.UserID,
		Type:      event.Type,
		Data:      event.Data,
		Timestamp: event.CreatedAt,
	}

	if err := s.kafkaProducer.SendEvent(ctx, kafkaEvent); err != nil {
		s.logger.Errorf("Failed to send event to Kafka: %v", err)
	}

	s.logger.Infof("Event created: %s", event.ID)
	return event, nil
}

func (s *EventService) GetEvent(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("event:%s", id.String())
	if cached, err := s.redisClient.Get(ctx, cacheKey); err == nil {
		var event models.Event
		if err := json.Unmarshal([]byte(cached), &event); err == nil {
			s.logger.Debugf("Event %s retrieved from cache", id)
			return &event, nil
		}
	}

	// Get from database
	event := &models.Event{}
	query := `SELECT id, user_id, type, data, created_at FROM events WHERE id = $1`

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID, &event.UserID, &event.Type, &event.Data, &event.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Cache the result
	s.cacheEvent(ctx, event)

	s.logger.Debugf("Event %s retrieved from database", id)
	return event, nil
}

func (s *EventService) ListEvents(ctx context.Context, page, limit int) (*models.EventListResponse, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM events`
	err := s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}

	// Get events
	query := `
		SELECT id, user_id, type, data, created_at 
		FROM events 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(&event.ID, &event.UserID, &event.Type, &event.Data, &event.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %v", err)
		}
		events = append(events, event)
	}

	return &models.EventListResponse{
		Events: events,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
}

func (s *EventService) ProcessEvents(consumer *kafka.Consumer) {
	s.logger.Info("Starting event processing...")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		event, err := consumer.ReadMessage(ctx)
		if err != nil {
			s.logger.Errorf("Failed to read message from Kafka: %v", err)
			cancel()
			time.Sleep(5 * time.Second)
			continue
		}

		// Process event in a goroutine for parallel processing
		go s.processEvent(event)

		cancel()
	}
}

func (s *EventService) processEvent(event models.KafkaEvent) {
	s.logger.Infof("Processing event: %s (type: %s)", event.ID, event.Type)

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	// Here you would implement your business logic for processing events
	// For example: sending notifications, updating analytics, etc.

	s.logger.Infof("Event processed successfully: %s", event.ID)
}

func (s *EventService) cacheEvent(ctx context.Context, event *models.Event) {
	cacheKey := fmt.Sprintf("event:%s", event.ID.String())
	eventData, err := json.Marshal(event)
	if err != nil {
		s.logger.Errorf("Failed to marshal event for cache: %v", err)
		return
	}

	if err := s.redisClient.Set(ctx, cacheKey, string(eventData), 30*time.Minute); err != nil {
		s.logger.Errorf("Failed to cache event: %v", err)
	}
}
