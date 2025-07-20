package models

import (
	"github.com/google/uuid"
	"time"
)

type NotificationStatus string

var (
	StatusSent     NotificationStatus = "sent"
	StatusFailed   NotificationStatus = "failed"
	StatusPending  NotificationStatus = "pending"
	StatusInFlight NotificationStatus = "in_flight"
)

type Notification struct {
	ID             uuid.UUID
	UserID         int
	Text           string
	RecipientPhone string
	Status         NotificationStatus
	Attempts       int
	NextRunAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
