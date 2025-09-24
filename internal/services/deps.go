package services

import (
	"context"
	"time"

	"highload-microservice/internal/models"
)

// RedisClient abstracts the subset of Redis methods used by services.
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

// KafkaProducer abstracts sending events to Kafka.
type KafkaProducer interface {
	SendEvent(ctx context.Context, event models.KafkaEvent) error
}
