package domain

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/apiservice/internal/bootstrap"
	"github.com/segmentio/kafka-go"
)

// KafkaFactory defines the contract for creating Kafka writers and performing health checks.
type KafkaFactory interface {
	Ping(ctx context.Context) error
	NewWriter(topic string, opts ...bootstrap.WriterOption) *kafka.Writer
}

// KafkaWriter abstracts the production of messages to a Kafka topic.
// Implementations should handle batching, retries, and context-based cancellations.
type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

// KafkaReader defines the interface for consuming messages from Kafka.
// Implementations should handle fetching and committing offsets.
type KafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}
