package shared

import (
	"context"
	"fmt"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/database"
)

const (
	queryCreateOrder = `
		INSERT INTO orders (cart_id, buyer_id, total_price, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	queryCreateItem = `
		INSERT INTO order_items (order_id, product_id, quantity, unit_price_snapshot, discount_snapshot)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
)

// Repository only talks to the database; see the layer rules in
// sdd/specs/infra-module-architecture.md. Read and status operations arrive
// with the order endpoints (Phase 3).
type Repository interface {
	// Create inserts the order and its items. Run it inside the caller's
	// transaction so the order is all-or-nothing.
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
}

type repository struct {
	db database.DBTX
}

func NewRepository(db database.DBTX) Repository {
	return &repository{db: db}
}

func wrap(op string, err error) error {
	return fmt.Errorf("order repository: %s: %w", op, err)
}

func (r *repository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	err := r.db.QueryRowContext(ctx, queryCreateOrder,
		order.CartID, order.BuyerID, order.TotalPrice, order.Status,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return domain.Order{}, wrap("create order", err)
	}

	items := make([]domain.Item, len(order.Items))
	for i, item := range order.Items {
		item.OrderID = order.ID
		err := r.db.QueryRowContext(ctx, queryCreateItem,
			item.OrderID, item.ProductID, item.Quantity, item.UnitPrice, item.DiscountPercentage,
		).Scan(&item.ID)
		if err != nil {
			return domain.Order{}, wrap("create order item", err)
		}
		items[i] = item
	}
	order.Items = items

	return order, nil
}
