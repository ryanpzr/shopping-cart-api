package adminactivitylog

import (
	"time"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/domain"
)

// ActivityLogEntry is one audited action. ID is the event id.
type ActivityLogEntry struct {
	ID         string         `json:"id"`
	UserID     int            `json:"user_id"`
	EventType  string         `json:"event_type"`
	Metadata   map[string]any `json:"metadata"`
	OccurredAt time.Time      `json:"occurred_at"`
}

type ActivityLogResponse struct {
	Data       []ActivityLogEntry `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

func toEntry(e domain.Entry) ActivityLogEntry {
	metadata := e.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	return ActivityLogEntry{
		ID:         e.EventID,
		UserID:     e.UserID,
		EventType:  e.EventType,
		Metadata:   metadata,
		OccurredAt: e.OccurredAt,
	}
}
