package bootstrap

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

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
func (kf *KafkaFactory) NewWriter(topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kf.Addrs...),
		Balancer: kf.Balancer,
		Topic:    topic,
	}
}

// NewReader creates a kafka.Reader for the specified topic and consumer group.
// The reader is configured with the factory’s broker addresses and a maximum
// message fetch size of 10MB.
func (kf *KafkaFactory) NewReader(topic string, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  kf.Addrs,
		GroupID:  groupID,
		Topic:    topic,
		MaxBytes: 10e6, // 10MB
	})
}
