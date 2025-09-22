package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
    Email     string `json:"email" binding:"required,email" validate:"required,email,email_domain,no_sql_injection,no_xss"`
	FirstName string `json:"first_name" binding:"required" validate:"required,min=1,max=100,safe_string,no_sql_injection,no_xss"`
	LastName  string `json:"last_name" binding:"required" validate:"required,min=1,max=100,safe_string,no_sql_injection,no_xss"`
}

type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" validate:"omitempty,email,email_domain,no_sql_injection,no_xss"`
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100,safe_string,no_sql_injection,no_xss"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100,safe_string,no_sql_injection,no_xss"`
}

type UserListResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}
