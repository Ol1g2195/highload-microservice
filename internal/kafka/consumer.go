package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"highload-microservice/internal/config"
	"highload-microservice/internal/models"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(cfg config.KafkaConfig) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{reader: reader}, nil
}

func (c *Consumer) ReadMessage(ctx context.Context) (models.KafkaEvent, error) {
	var event models.KafkaEvent

	message, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return event, fmt.Errorf("failed to read message: %w", err)
	}

	if err := json.Unmarshal(message.Value, &event); err != nil {
		return event, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return event, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

