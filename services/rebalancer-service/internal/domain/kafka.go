package domain

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// KafkaFactory defines the contract for creating Kafka writers and performing health checks.
type KafkaFactory interface {
	Ping(ctx context.Context) error
	NewWriter(topic string) *kafka.Writer
}

// KafkaWriter abstracts the production of messages to a Kafka topic.
// Implementations should handle batching, retries, and context-based cancellations.
type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}
