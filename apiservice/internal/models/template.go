package models

import "time"

// Template represents a message template created by a user.
type Template struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	Body         string    `json:"body"`
	CreationTime time.Time `json:"creationTime"`
	UpdateTime   time.Time `json:"updateTime"`
}
