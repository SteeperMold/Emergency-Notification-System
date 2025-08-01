package models

// SlimContact represents the minimal information needed to send a notification.
// It omits database metadata and user associations
type SlimContact struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}
