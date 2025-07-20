package domain

import (
	"context"
	"github.com/segmentio/kafka-go"
)

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
