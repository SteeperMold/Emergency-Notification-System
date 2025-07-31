package bootstrap

import (
	"github.com/segmentio/kafka-go"
	"time"
)

type WriterOption func(w *kafka.Writer)

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
