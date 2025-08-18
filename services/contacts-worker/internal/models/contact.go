package models

import "time"

// Contact represents a user's contact information stored in the system.
type Contact struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	CreationTime time.Time `json:"creationTime"`
	UpdateTime   time.Time `json:"updateTime"`
}
