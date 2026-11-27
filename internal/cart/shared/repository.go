package shared

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/database"
)

// Repository only talks to the database and returns domain errors or wrapped
// errors; usecases own the rules (see sdd/specs/infra-module-architecture.md).
type Repository interface {
	// FindByBuyer returns the buyer's cart with the given status, with its items.
	FindByBuyer(ctx context.Context, buyerID int, status string) (domain.Cart, error)
	// LockByBuyer is FindByBuyer with the cart row locked (SELECT … FOR UPDATE).
	// It must run inside a transaction.
	LockByBuyer(ctx context.Context, buyerID int, status string) (domain.Cart, error)
	// Create inserts an empty cart. It returns domain.ErrActiveCartExists when
	// another request created the buyer's active cart first.
	Create(ctx context.Context, buyerID int, status string) (domain.Cart, error)
	UpdateStatus(ctx context.Context, cartID int, status string) error
	AddItem(ctx context.Context, item domain.Item) (domain.Item, error)
	// UpdateItem writes the quantity and the price snapshot.
	UpdateItem(ctx context.Context, item domain.Item) error
	RemoveItem(ctx context.Context, cartID, itemID int) error
}

type repository struct {
	db database.DBTX
}

func NewRepository(db database.DBTX) Repository {
	return &repository{db: db}
}

func wrap(op string, err error) error {
	return fmt.Errorf("cart repository: %s: %w", op, err)
}

func scanCart(row *sql.Row) (domain.Cart, error) {
	var c domain.Cart
	err := row.Scan(&c.ID, &c.BuyerID, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *repository) FindByBuyer(ctx context.Context, buyerID int, status string) (domain.Cart, error) {
	return r.findByBuyer(ctx, queryFindByBuyer, buyerID, status)
}

func (r *repository) LockByBuyer(ctx context.Context, buyerID int, status string) (domain.Cart, error) {
	return r.findByBuyer(ctx, queryLockByBuyer, buyerID, status)
}

func (r *repository) findByBuyer(ctx context.Context, query string, buyerID int, status string) (domain.Cart, error) {
	cart, err := scanCart(r.db.QueryRowContext(ctx, query, buyerID, status))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Cart{}, domain.ErrCartNotFound
	}
	if err != nil {
		return domain.Cart{}, wrap("find by buyer", err)
	}

	cart.Items, err = r.listItems(ctx, cart.ID)
	if err != nil {
		return domain.Cart{}, err
	}
	return cart, nil
}

func (r *repository) listItems(ctx context.Context, cartID int) ([]domain.Item, error) {
	rows, err := r.db.QueryContext(ctx, queryListItems, cartID)
	if err != nil {
		return nil, wrap("list items", err)
	}
	defer func() { _ = rows.Close() }()

	items := []domain.Item{}
	for rows.Next() {
		var i domain.Item
		if err := rows.Scan(&i.ID, &i.CartID, &i.ProductID, &i.Title, &i.Quantity, &i.PriceSnapshot, &i.CreatedAt); err != nil {
			return nil, wrap("scan item", err)
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, wrap("iterate items", err)
	}
	return items, nil
}

func (r *repository) Create(ctx context.Context, buyerID int, status string) (domain.Cart, error) {
	cart, err := scanCart(r.db.QueryRowContext(ctx, queryCreate, buyerID, status))
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == constraintOneActiveCart {
		return domain.Cart{}, domain.ErrActiveCartExists
	}
	if err != nil {
		return domain.Cart{}, wrap("create", err)
	}
	cart.Items = []domain.Item{}
	return cart, nil
}

func (r *repository) UpdateStatus(ctx context.Context, cartID int, status string) error {
	result, err := r.db.ExecContext(ctx, queryUpdateStatus, status, cartID)
	if err != nil {
		return wrap("update status", err)
	}
	return requireAffected(result, "update status", domain.ErrCartNotFound)
}

func (r *repository) AddItem(ctx context.Context, item domain.Item) (domain.Item, error) {
	err := r.db.QueryRowContext(ctx, queryAddItem,
		item.CartID, item.ProductID, item.Quantity, item.PriceSnapshot,
	).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return domain.Item{}, wrap("add item", err)
	}
	return item, nil
}

func (r *repository) UpdateItem(ctx context.Context, item domain.Item) error {
	result, err := r.db.ExecContext(ctx, queryUpdateItem, item.Quantity, item.PriceSnapshot, item.ID, item.CartID)
	if err != nil {
		return wrap("update item", err)
	}
	return requireAffected(result, "update item", domain.ErrItemNotFound)
}

func (r *repository) RemoveItem(ctx context.Context, cartID, itemID int) error {
	result, err := r.db.ExecContext(ctx, queryRemoveItem, itemID, cartID)
	if err != nil {
		return wrap("remove item", err)
	}
	return requireAffected(result, "remove item", domain.ErrItemNotFound)
}

func requireAffected(result sql.Result, op string, notFound error) error {
	n, err := result.RowsAffected()
	if err != nil {
		return wrap(op, err)
	}
	if n == 0 {
		return notFound
	}
	return nil
}
