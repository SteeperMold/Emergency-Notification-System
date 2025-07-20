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

type SlimContact struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func ToSlim(contacts []*Contact) []*SlimContact {
	slim := make([]*SlimContact, len(contacts))
	for i, c := range contacts {
		slim[i] = &SlimContact{
			Name:  c.Name,
			Phone: c.Phone,
		}
	}
	return slim
}
