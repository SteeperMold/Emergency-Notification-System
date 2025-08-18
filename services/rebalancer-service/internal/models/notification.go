package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a single notification record in the system.
type Notification struct {
	ID             uuid.UUID
	UserID         int
	Text           string
	RecipientPhone string
	Status         string
	Attempts       int
	NextRunAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
