package models

import "time"

// User represents a registered user in the system.
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreationTime time.Time `json:"creationTime"`
}
