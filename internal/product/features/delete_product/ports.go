package deleteproduct

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

type Usecase interface {
	Delete(ctx context.Context, productID, userID int, role string) error
}

// OrderChecker tells whether a product appears in orders with given statuses.
// Implemented by the order repository (internal/order/shared).
type OrderChecker interface {
	HasOrdersWithStatus(ctx context.Context, productID int, statuses []string) (bool, error)
}

// TxRepositories are repositories bound to the same transaction.
type TxRepositories struct {
	Products shared.Repository
	Orders   OrderChecker
	Outbox   outbox.Writer
}

// UnitOfWork runs fn inside one transaction: it commits when fn returns nil
// and rolls back otherwise. Usecases depend on this port, so they stay
// testable without a database.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos TxRepositories) error) error
}
