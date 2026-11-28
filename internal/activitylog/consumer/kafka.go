package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConfig is the Kafka side of the consumer.
type KafkaConfig struct {
	Brokers  []string
	Topic    string
	GroupID  string
	DLQTopic string
}

// NewReader returns the consumer group reader. A new group starts from the
// beginning of the topic; the offset is committed explicitly by Run (a zero
// CommitInterval makes CommitMessages synchronous).
func NewReader(cfg KafkaConfig) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		MinBytes:    1,
		MaxBytes:    10 << 20,
		MaxWait:     time.Second,
		StartOffset: kafka.FirstOffset,
	})
}

// DLQPublisher writes malformed messages to the dead-letter topic.
type DLQPublisher struct {
	writer *kafka.Writer
	source string
}

func NewDLQPublisher(cfg KafkaConfig) *DLQPublisher {
	return &DLQPublisher{
		source: cfg.Topic,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.DLQTopic,
			RequiredAcks: kafka.RequireAll,
			Balancer:     &kafka.Hash{},
		},
	}
}

// Publish keeps the original key and value and records why it was rejected.
func (p *DLQPublisher) Publish(ctx context.Context, key, value []byte, reason string) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Headers: []kafka.Header{
			{Key: "error", Value: []byte(reason)},
			{Key: "source-topic", Value: []byte(p.source)},
		},
	})
	if err != nil {
		return fmt.Errorf("write to %s: %w", p.writer.Topic, err)
	}
	return nil
}

func (p *DLQPublisher) Close() error { return p.writer.Close() }
