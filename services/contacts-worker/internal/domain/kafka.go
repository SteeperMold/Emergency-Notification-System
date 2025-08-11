package domain

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// KafkaFactory defines the contract for creating Kafka writers and performing health checks.
type KafkaFactory interface {
	Ping(ctx context.Context) error
	NewReader(topic string, groupID string) *kafka.Reader
}

// KafkaReader defines the interface for consuming messages from Kafka.
// Implementations should handle fetching and committing offsets.
type KafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}
