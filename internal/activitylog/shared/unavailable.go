package shared

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/domain"
)

// DefaultDatabase is the MongoDB database used when MONGO_DB is not set.
const DefaultDatabase = "activity_log"

type noStore struct{}

// NewUnavailableRepository is the store of an API that has no MongoDB
// configured: every call fails with domain.ErrStoreUnavailable, so the read
// endpoint answers 503 while the rest of the API works normally.
func NewUnavailableRepository() Repository { return noStore{} }

func (noStore) Upsert(context.Context, domain.Entry) error { return domain.ErrStoreUnavailable }

func (noStore) FindByUser(context.Context, int, string, int, int) ([]domain.Entry, int64, error) {
	return nil, 0, domain.ErrStoreUnavailable
}

func (noStore) EnsureIndexes(context.Context) error { return domain.ErrStoreUnavailable }
