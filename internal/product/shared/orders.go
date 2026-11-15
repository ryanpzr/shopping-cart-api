package shared

import (
	"context"

	"github.com/lib/pq"
)

// OrderRepository answers questions about the orders that reference a product.
// It needs the orders and order_items tables from module 05 (Phase 3).
type OrderRepository struct {
	db DBTX
}

func NewOrderRepository(db DBTX) *OrderRepository {
	return &OrderRepository{db: db}
}

// HasOrdersWithStatus reports whether the product is in an order whose status
// is one of statuses. Which statuses matter is decided by the caller.
func (r *OrderRepository) HasOrdersWithStatus(ctx context.Context, productID int, statuses []string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, queryHasOrdersWithStatus, productID, pq.Array(statuses)).Scan(&exists)
	if err != nil {
		return false, wrap("has orders with status", err)
	}
	return exists, nil
}

// NoOrders stands in for OrderRepository until the order tables exist
// (Phase 3): with no orders, no product can be referenced by one.
// TODO(module-05): wire NewOrderRepository in internal/app/setup/setup.go and delete this type.
type NoOrders struct{}

func (NoOrders) HasOrdersWithStatus(context.Context, int, []string) (bool, error) {
	return false, nil
}
