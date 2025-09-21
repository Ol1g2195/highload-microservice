package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"`
	Data      string    `json:"data" db:"data"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateEventRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required" validate:"required,uuid"`
	Type   string    `json:"type" binding:"required" validate:"required,min=1,max=50,safe_string,no_sql_injection,no_xss"`
	Data   string    `json:"data" binding:"required" validate:"required,min=1,max=1000,safe_string,no_sql_injection,no_xss"`
}

type EventListResponse struct {
	Events []Event `json:"events"`
	Total  int     `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}

type KafkaEvent struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"`
	Data      string    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}
