// Package consumer moves audit events from Kafka into the MongoDB audit store.
//
// Delivery is at-least-once: the offset is committed only after the entry is
// stored, and storing the same event_id twice is a no-op, so a redelivery after
// a crash does no harm.
package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

// Store is the audit store the consumer writes to.
type Store interface {
	Upsert(ctx context.Context, entry domain.Entry) error
}

// DeadLetters receives the messages that can never be processed.
type DeadLetters interface {
	Publish(ctx context.Context, key, value []byte, reason string) error
}

// Backoff is the wait between retries of a failing message: it doubles from
// Initial up to Max.
type Backoff struct {
	Initial time.Duration
	Max     time.Duration
}

func (b Backoff) delay(attempt int) time.Duration {
	d := b.Initial
	for i := 0; i < attempt && d < b.Max; i++ {
		d *= 2
	}
	if d > b.Max {
		d = b.Max
	}
	return d
}

// Handler processes one Kafka message at a time.
type Handler struct {
	store   Store
	dlq     DeadLetters
	backoff Backoff
	now     func() time.Time
	wait    func(ctx context.Context, d time.Duration) error
}

func NewHandler(store Store, dlq DeadLetters, backoff Backoff) *Handler {
	return &Handler{store: store, dlq: dlq, backoff: backoff, now: time.Now, wait: sleep}
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Handle processes the message once. A nil result means the offset may be
// committed: the entry was stored (or already existed), or the message is
// malformed and now sits in the dead-letter topic. Any other error is
// transient: the offset must not be committed.
func (h *Handler) Handle(ctx context.Context, key, value []byte) error {
	entry, err := decode(value, h.now())
	if err != nil {
		if dlqErr := h.dlq.Publish(ctx, key, value, err.Error()); dlqErr != nil {
			return fmt.Errorf("publish malformed message to the dead-letter topic: %w", dlqErr)
		}
		return nil
	}
	return h.store.Upsert(ctx, entry)
}

// Process calls Handle until it succeeds, waiting with exponential backoff
// between attempts. It returns only when the message is handled or ctx ends;
// in the second case the offset must not be committed.
func (h *Handler) Process(ctx context.Context, key, value []byte, onRetry func(attempt int, err error)) error {
	for attempt := 0; ; attempt++ {
		err := h.Handle(ctx, key, value)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if onRetry != nil {
			onRetry(attempt, err)
		}
		if err := h.wait(ctx, h.backoff.delay(attempt)); err != nil {
			return err
		}
	}
}

// errMalformed marks a message that can never become an entry.
var errMalformed = errors.New("malformed audit event")

// decode turns the Kafka message value into an audit entry.
func decode(value []byte, ingestedAt time.Time) (domain.Entry, error) {
	var msg outbox.Message
	if err := json.Unmarshal(value, &msg); err != nil {
		return domain.Entry{}, fmt.Errorf("%w: not a JSON event: %v", errMalformed, err)
	}
	if _, err := uuid.Parse(msg.EventID); err != nil {
		return domain.Entry{}, fmt.Errorf("%w: event_id is not a UUID", errMalformed)
	}
	if msg.UserID < 1 {
		return domain.Entry{}, fmt.Errorf("%w: user_id is missing", errMalformed)
	}
	if !outbox.ValidType(msg.EventType) {
		return domain.Entry{}, fmt.Errorf("%w: unknown event_type %q", errMalformed, msg.EventType)
	}
	if msg.OccurredAt.IsZero() {
		return domain.Entry{}, fmt.Errorf("%w: occurred_at is missing", errMalformed)
	}

	return domain.Entry{
		EventID:    msg.EventID,
		UserID:     msg.UserID,
		EventType:  msg.EventType,
		Metadata:   msg.Metadata,
		OccurredAt: msg.OccurredAt.UTC(),
		IngestedAt: ingestedAt.UTC(),
	}, nil
}
