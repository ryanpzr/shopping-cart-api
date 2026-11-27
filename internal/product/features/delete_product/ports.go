package deleteproduct

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Usecase interface {
	Delete(ctx context.Context, productID, userID int, role string) error
}

type Repository = shared.Repository

// OrderChecker tells whether a product appears in orders with given statuses.
// Implemented by the order repository (internal/order/shared).
type OrderChecker interface {
	HasOrdersWithStatus(ctx context.Context, productID int, statuses []string) (bool, error)
}
