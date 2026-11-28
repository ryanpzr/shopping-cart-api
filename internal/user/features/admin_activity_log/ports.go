package adminactivitylog

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/domain"
)

type Usecase interface {
	// GetActivityLog returns one page of the user's audit entries, newest
	// first, optionally filtered by event type.
	GetActivityLog(ctx context.Context, userID, page, limit int, eventType string) (ActivityLogResponse, error)
}

// Repository is the read side of the audit store (MongoDB).
type Repository interface {
	FindByUser(ctx context.Context, userID int, eventType string, limit, offset int) ([]domain.Entry, int64, error)
}
