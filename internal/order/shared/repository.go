package shared

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"

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

	queryHasOrdersWithStatus = `
		SELECT EXISTS (
			SELECT 1
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE oi.product_id = $1
			  AND o.status = ANY($2)
		)
	`

	// Name of the unique constraint that allows one order per cart.
	constraintOneOrderPerCart = "orders_cart_id_key"

	pgUniqueViolation = "23505"
)

// Repository only talks to the database; see the layer rules in
// sdd/specs/infra-module-architecture.md. Read and status operations arrive
// with the order endpoints (Phase 3).
type Repository interface {
	// Create inserts the order and its items. Run it inside the caller's
	// transaction so the order is all-or-nothing. It returns
	// domain.ErrCartAlreadyOrdered when the cart already has an order.
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	// HasOrdersWithStatus reports whether the product is in an order whose
	// status is one of statuses. Which statuses matter is decided by the caller.
	HasOrdersWithStatus(ctx context.Context, productID int, statuses []string) (bool, error)
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
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == pgUniqueViolation && pqErr.Constraint == constraintOneOrderPerCart {
		return domain.Order{}, domain.ErrCartAlreadyOrdered
	}
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

func (r *repository) HasOrdersWithStatus(ctx context.Context, productID int, statuses []string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, queryHasOrdersWithStatus, productID, pq.Array(statuses)).Scan(&exists)
	if err != nil {
		return false, wrap("has orders with status", err)
	}
	return exists, nil
}
