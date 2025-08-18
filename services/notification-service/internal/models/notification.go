package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationStatus represents the lifecycle state of a notification
type NotificationStatus string

var (
	// StatusSent indicates the notification was delivered successfully
	StatusSent NotificationStatus = "sent"
	// StatusFailed indicates the notification has permanently failed
	StatusFailed NotificationStatus = "failed"
	// StatusPending indicates the notification is scheduled but not yet attempted
	StatusPending NotificationStatus = "pending"
	// StatusInFlight indicates the notification is currently being sent
	StatusInFlight NotificationStatus = "in_flight"
)

// Notification captures all relevant data for a single SMS notification task
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
