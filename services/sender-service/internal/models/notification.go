package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a message that is scheduled to be sent to a recipient via SMS.
// It contains metadata about the user, status, retry attempts, scheduling, and timestamps.
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
