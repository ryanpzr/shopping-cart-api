package adminactivitylog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

// GetActivityLog implements BR-ADMIN-006. It reads from MongoDB, which is fed
// asynchronously, so a very recent action may not be listed yet.
func (u *usecase) GetActivityLog(ctx context.Context, userID, page, limit int, eventType string) (ActivityLogResponse, error) {
	if eventType != "" && !outbox.ValidType(eventType) {
		return ActivityLogResponse{}, apperrors.NewBadRequest(
			fmt.Sprintf("event_type must be one of: %s", strings.Join(outbox.EventTypes, ", ")))
	}

	entries, total, err := u.repo.FindByUser(ctx, userID, eventType, limit, (page-1)*limit)
	if errors.Is(err, domain.ErrStoreUnavailable) {
		// The cause stays in the log; the client only gets a fixed message.
		log.Printf("activity log: %v", err)
		return ActivityLogResponse{}, apperrors.NewServiceUnavailable("activity log is temporarily unavailable")
	}
	if err != nil {
		return ActivityLogResponse{}, apperrors.NewInternalServer(err.Error())
	}

	data := make([]ActivityLogEntry, 0, len(entries))
	for _, e := range entries {
		data = append(data, toEntry(e))
	}
	return ActivityLogResponse{
		Data:       data,
		Total:      int(total),
		Page:       page,
		Limit:      limit,
		TotalPages: (int(total) + limit - 1) / limit,
	}, nil
}
