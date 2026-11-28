// Package mocks provides a testify mock of the outbox Writer, shared by the
// usecase tests that emit audit events.
package mocks

import (
	"context"
	"reflect"

	"github.com/stretchr/testify/mock"

	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

type Writer struct{ mock.Mock }

var _ outbox.Writer = (*Writer)(nil)

func (m *Writer) Insert(ctx context.Context, e outbox.Event) error {
	return m.Called(ctx, e).Error(0)
}

// Matches matches an event by type, user and metadata (deep equality). The id
// and the time are random, so tests never compare them, but the event must be
// valid.
func Matches(eventType string, userID int, metadata map[string]any) any {
	return mock.MatchedBy(func(e outbox.Event) bool {
		if e.Type != eventType || e.UserID != userID || e.Validate() != nil {
			return false
		}
		got := e.Metadata
		if got == nil {
			got = map[string]any{}
		}
		if metadata == nil {
			metadata = map[string]any{}
		}
		return reflect.DeepEqual(got, metadata)
	})
}
