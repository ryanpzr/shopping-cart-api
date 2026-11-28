// Package domain holds the audit entry stored by the activity log.
package domain

import (
	"errors"
	"time"
)

// ErrStoreUnavailable is returned by the audit store when it cannot be
// reached. Usecases decide how to present it (the read endpoint answers 503).
var ErrStoreUnavailable = errors.New("activity log store unavailable")

// Entry is one audited action, as stored in MongoDB. EventID is the id the
// API generated for the outbox event: it makes the consumer idempotent.
type Entry struct {
	EventID    string
	UserID     int
	EventType  string
	Metadata   map[string]any
	OccurredAt time.Time
	IngestedAt time.Time
}
