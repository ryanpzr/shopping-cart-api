package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/ryanpzr/shopping-cart-api/pkg/database"
)

const queryInsert = `
	INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload)
	VALUES ($1, $2, $3, $4, $5::jsonb)
`

// Event is one audited action. UserID is the user who did it.
type Event struct {
	ID         uuid.UUID
	UserID     int
	Type       string
	Metadata   map[string]any
	OccurredAt time.Time
}

// New builds an event with a fresh id and the current time.
func New(userID int, eventType string, metadata map[string]any) Event {
	return Event{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       eventType,
		Metadata:   metadata,
		OccurredAt: time.Now().UTC(),
	}
}

// Message is the JSON document that travels through Kafka: the outbox payload
// column, the Kafka message value and, in the end, the MongoDB document.
type Message struct {
	EventID    string         `json:"event_id"`
	UserID     int            `json:"user_id"`
	EventType  string         `json:"event_type"`
	Metadata   map[string]any `json:"metadata"`
	OccurredAt time.Time      `json:"occurred_at"`
}

// Validate rejects events that could not be audited meaningfully.
func (e Event) Validate() error {
	if e.ID == uuid.Nil {
		return errors.New("outbox: event id is required")
	}
	if e.UserID < 1 {
		return errors.New("outbox: user id is required")
	}
	if !ValidType(e.Type) {
		return fmt.Errorf("outbox: unknown event type %q", e.Type)
	}
	if e.OccurredAt.IsZero() {
		return errors.New("outbox: occurred_at is required")
	}
	return nil
}

// Message converts the event to its wire format. A nil Metadata becomes {}.
func (e Event) Message() Message {
	metadata := e.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	return Message{
		EventID:    e.ID.String(),
		UserID:     e.UserID,
		EventType:  e.Type,
		Metadata:   metadata,
		OccurredAt: e.OccurredAt.UTC(),
	}
}

// Writer inserts events through a database handle. Bind it to the *sql.Tx of
// the business change (database.DBTX accepts both *sql.DB and *sql.Tx).
type Writer interface {
	// Insert writes the event. It must run in the same transaction as the
	// business change it describes: if it fails, the caller must roll back.
	Insert(ctx context.Context, e Event) error
}

type writer struct {
	db database.DBTX
}

// NewWriter returns a Writer on the given handle.
func NewWriter(db database.DBTX) Writer {
	return &writer{db: db}
}

func (w *writer) Insert(ctx context.Context, e Event) error {
	if err := e.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(e.Message())
	if err != nil {
		return fmt.Errorf("outbox: encode payload: %w", err)
	}

	_, err = w.db.ExecContext(ctx, queryInsert,
		e.ID, AggregateType(e.Type), strconv.Itoa(e.UserID), e.Type, string(payload))
	if err != nil {
		return fmt.Errorf("outbox: insert event: %w", err)
	}
	return nil
}
