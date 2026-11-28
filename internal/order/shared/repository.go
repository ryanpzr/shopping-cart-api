package shared

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/database"
)

// Repository only talks to the database; see the layer rules in
// sdd/specs/infra-module-architecture.md. It returns domain errors or wrapped
// errors and holds no business rule: usecases decide who may see or change an
// order and which status changes are allowed.
type Repository interface {
	// Create inserts the order and its items. Run it inside the caller's
	// transaction so the order is all-or-nothing. It returns
	// domain.ErrCartAlreadyOrdered when the cart already has an order.
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	// FindByBuyer returns one page of the buyer's orders, newest first, with
	// their items, and the total number of orders of the buyer.
	FindByBuyer(ctx context.Context, buyerID, limit, offset int) ([]domain.Order, int, error)
	// FindByID returns the order with its items, or domain.ErrOrderNotFound.
	FindByID(ctx context.Context, id int) (domain.Order, error)
	// LockByID is FindByID with the order row locked (SELECT … FOR UPDATE).
	// It must run inside a transaction.
	LockByID(ctx context.Context, id int) (domain.Order, error)
	UpdateStatus(ctx context.Context, id int, status string) error
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

type scanner interface {
	Scan(dest ...any) error
}

func scanOrder(s scanner) (domain.Order, error) {
	var o domain.Order
	err := s.Scan(&o.ID, &o.CartID, &o.BuyerID, &o.TotalPrice, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	return o, err
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

func (r *repository) FindByBuyer(ctx context.Context, buyerID, limit, offset int) ([]domain.Order, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, queryCountByBuyer, buyerID).Scan(&total); err != nil {
		return nil, 0, wrap("count by buyer", err)
	}

	orders, err := r.listByBuyer(ctx, buyerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := r.attachItems(ctx, orders); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// listByBuyer reads the page and closes its rows before the caller runs the
// next query, which matters inside a transaction (one connection).
func (r *repository) listByBuyer(ctx context.Context, buyerID, limit, offset int) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, queryListByBuyer, buyerID, limit, offset)
	if err != nil {
		return nil, wrap("list by buyer", err)
	}
	defer func() { _ = rows.Close() }()

	orders := make([]domain.Order, 0, limit)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, wrap("scan order", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, wrap("iterate orders", err)
	}
	return orders, nil
}

func (r *repository) FindByID(ctx context.Context, id int) (domain.Order, error) {
	return r.findByID(ctx, queryFindByID, id)
}

func (r *repository) LockByID(ctx context.Context, id int) (domain.Order, error) {
	return r.findByID(ctx, queryLockByID, id)
}

func (r *repository) findByID(ctx context.Context, query string, id int) (domain.Order, error) {
	order, err := scanOrder(r.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, wrap("find by id", err)
	}

	orders := []domain.Order{order}
	if err := r.attachItems(ctx, orders); err != nil {
		return domain.Order{}, err
	}
	return orders[0], nil
}

// attachItems loads the items of all orders with one query and sets them on
// each order (an empty slice when an order has none).
func (r *repository) attachItems(ctx context.Context, orders []domain.Order) error {
	if len(orders) == 0 {
		return nil
	}
	index := make(map[int]int, len(orders))
	ids := make([]int, 0, len(orders))
	for i := range orders {
		orders[i].Items = []domain.Item{}
		index[orders[i].ID] = i
		ids = append(ids, orders[i].ID)
	}

	rows, err := r.db.QueryContext(ctx, queryItemsByOrders, pq.Array(ids))
	if err != nil {
		return wrap("list items", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Title, &item.Quantity, &item.UnitPrice, &item.DiscountPercentage); err != nil {
			return wrap("scan item", err)
		}
		i := index[item.OrderID]
		orders[i].Items = append(orders[i].Items, item)
	}
	if err := rows.Err(); err != nil {
		return wrap("iterate items", err)
	}
	return nil
}

func (r *repository) UpdateStatus(ctx context.Context, id int, status string) error {
	result, err := r.db.ExecContext(ctx, queryUpdateStatus, status, id)
	if err != nil {
		return wrap("update status", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return wrap("update status", err)
	}
	if n == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *repository) HasOrdersWithStatus(ctx context.Context, productID int, statuses []string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, queryHasOrdersWithStatus, productID, pq.Array(statuses)).Scan(&exists)
	if err != nil {
		return false, wrap("has orders with status", err)
	}
	return exists, nil
}
