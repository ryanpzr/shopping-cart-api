package consumer

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// Source is the Kafka consumer group reader (a *kafka.Reader).
type Source interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

// Run reads messages one by one: it handles each, and only then commits its
// offset. It returns nil when ctx is cancelled (a clean shutdown) and an error
// when the source or the commit fails, so the process restarts and resumes
// from the last committed offset.
func Run(ctx context.Context, src Source, h *Handler) error {
	for {
		msg, err := src.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		err = h.Process(ctx, msg.Key, msg.Value, func(attempt int, err error) {
			log.Printf("activity consumer: partition %d offset %d failed (attempt %d), retrying: %v",
				msg.Partition, msg.Offset, attempt+1, err)
		})
		if err != nil {
			return nil // ctx ended while retrying: the offset stays uncommitted
		}

		if err := src.CommitMessages(ctx, msg); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("commit offset %d of partition %d: %w", msg.Offset, msg.Partition, err)
		}
	}
}
