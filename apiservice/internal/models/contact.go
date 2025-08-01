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

// SlimContact contains only the minimal fields (Name and Phone)
// needed when sending contact data to other services or clients.
type SlimContact struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// ToSlim transforms a slice of full Contact pointers into a slice
// of SlimContact pointers, dropping all metadata and IDs.
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
