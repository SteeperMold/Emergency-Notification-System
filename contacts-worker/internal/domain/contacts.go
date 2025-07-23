package domain

import (
	"context"
	"github.com/SteeperMold/Emergency-Notification-System/contacts-worker/internal/models"
)

// ContactsService defines the interface for processing contact upload tasks.
// Implementations should fetch the file from storage, process its contents,
// and return the number of successfully processed records.
type ContactsService interface {
	ProcessFile(ctx context.Context, task *Task) (processedContacts int, err error)
}

type ContactsRepository interface {
	SaveContacts(ctx context.Context, contacts []*models.Contact) error
}

// Task represents a job to load contacts from an S3 object for a specific user.
type Task struct {
	UserID int    `json:"userID"`
	S3Key  string `json:"s3Key"`
}
