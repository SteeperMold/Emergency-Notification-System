package bootstrap

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// WriterOption configures optional parameters for a Kafka writer.
type WriterOption func(w *kafka.Writer)

// WithBatchTimeout returns a WriterOption that sets a custom
// batch timeout for flushing messages to Kafka.
func WithBatchTimeout(d time.Duration) WriterOption {
	return func(w *kafka.Writer) {
		w.BatchTimeout = d
	}
}

// KafkaFactory produces configured Kafka writers.
// It encapsulates address list and load-balancing strategy.
type KafkaFactory struct {
	Addrs    []string
	Balancer kafka.Balancer
}

// NewKafkaFactory constructs a KafkaFactory given a KafkaConfig.
// Uses LeastBytes as the default partition balancing strategy.
func NewKafkaFactory(kafkaConfig *KafkaConfig) *KafkaFactory {
	return &KafkaFactory{
		Addrs:    kafkaConfig.KafkaAddrs,
		Balancer: &kafka.LeastBytes{},
	}
}

// Ping method makes sure that at least one broker is available.
func (kf *KafkaFactory) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var lastErr error
	for _, addr := range kf.Addrs {
		conn, err := kafka.DialContext(ctx, "tcp", addr)
		if err != nil {
			lastErr = err
			continue
		}

		err = conn.Close()
		if err != nil {
			return err
		}

		return nil
	}

	return lastErr
}

// NewWriter creates a kafka.Writer for the specified topic.
// The returned writer uses the factory's broker addresses and balancer.
func (kf *KafkaFactory) NewWriter(topic string, opts ...WriterOption) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(kf.Addrs...),
		Topic:    topic,
		Balancer: kf.Balancer,
	}
	for _, opt := range opts {
		opt(writer)
	}
	return writer
}
