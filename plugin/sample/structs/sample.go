package structs

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrNotFound     = errors.New("sample not found")
	ErrInvalidInput = errors.New("invalid input")
)

// Sample represents a sample entity
type Sample struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateSampleInput represents input for creating a sample
type CreateSampleInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateSampleInput represents input for updating a sample
type UpdateSampleInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}
