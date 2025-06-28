package domain

import (
	"context"
	"io"
)

// LoadContactsService defines the interface for services that handle
// uploading contact files and initiating their asynchronous processing.
// Implementations should store the file payload and enqueue a processing task.
type LoadContactsService interface {
	ProcessUpload(ctx context.Context, userID int, filename string, payload io.ReadSeeker) error
}

// LoadContactsTask represents the message payload published to Kafka
// for initiating contact file processing.
type LoadContactsTask struct {
	S3Key  string `json:"s3Key"`
	UserID int    `json:"userID"`
}
