package models

import (
	"github.com/google/uuid"
	"time"
)

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
