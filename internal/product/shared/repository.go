package shared

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

// DBTX is satisfied by both *sql.DB and *sql.Tx, so the repository can run
// inside a transaction owned by the caller (e.g. checkout, outbox events).
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type ProductFilters struct {
	Category *string
	MinPrice *float64
	MaxPrice *float64
	Search   *string
}

// ProductUpdate holds the optional fields of a partial update.
// Nil fields keep their current value.
type ProductUpdate struct {
	Photo              *string
	Title              *string
	Description        *string
	Price              *float64
	DiscountPercentage *int
	Quantity           *int
	Category           *string
}

type Repository interface {
	FindAll(ctx context.Context, filters ProductFilters, limit, offset int) ([]domain.Product, int, error)
	FindByID(ctx context.Context, id int) (*domain.Product, error)
	Create(ctx context.Context, p domain.Product) (domain.Product, error)
	Update(ctx context.Context, id int, u ProductUpdate) (domain.Product, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	Delete(ctx context.Context, id int) error
	HasActiveOrders(ctx context.Context, productID int) (bool, error)
}

type repository struct {
	db DBTX
}

func NewRepository(db DBTX) Repository {
	return &repository{db: db}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(s scanner) (domain.Product, error) {
	var p domain.Product
	err := s.Scan(
		&p.ID, &p.SellerID, &p.Photo, &p.Title, &p.Description,
		&p.Price, &p.DiscountPercentage, &p.Quantity,
		&p.Status, &p.Category, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func internal(op string, err error) error {
	return apperrors.NewInternalServer(fmt.Sprintf("product repository: %s: %v", op, err))
}

func (r *repository) FindAll(ctx context.Context, filters ProductFilters, limit, offset int) ([]domain.Product, int, error) {
	conditions := []string{"status = 'active'"}
	var args []any

	addCondition := func(format string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(format, len(args)))
	}
	if filters.Category != nil {
		addCondition("category = $%d", *filters.Category)
	}
	if filters.MinPrice != nil {
		addCondition("price >= $%d", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		addCondition("price <= $%d", *filters.MaxPrice)
	}
	if filters.Search != nil {
		addCondition("LOWER(title) LIKE $%d", "%"+strings.ToLower(*filters.Search)+"%")
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	var total int
	countQuery := "SELECT COUNT(*) FROM products " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, internal("count", err)
	}

	listQuery := fmt.Sprintf(`
		SELECT %s
		FROM products %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, productColumns, where, len(args)+1, len(args)+2)

	rows, err := r.db.QueryContext(ctx, listQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, internal("list", err)
	}
	defer func() { _ = rows.Close() }()

	products := make([]domain.Product, 0, limit)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, internal("scan", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, internal("iterate", err)
	}

	return products, total, nil
}

func (r *repository) FindByID(ctx context.Context, id int) (*domain.Product, error) {
	p, err := scanProduct(r.db.QueryRowContext(ctx, queryFindByID, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFound("product not found")
	}
	if err != nil {
		return nil, internal("find by id", err)
	}
	return &p, nil
}

func (r *repository) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	created, err := scanProduct(r.db.QueryRowContext(ctx, queryCreate,
		p.SellerID, p.Photo, p.Title, p.Description,
		p.Price, p.DiscountPercentage, p.Quantity, p.Status, p.Category,
	))
	if err != nil {
		return domain.Product{}, internal("create", err)
	}
	return created, nil
}

func (r *repository) Update(ctx context.Context, id int, u ProductUpdate) (domain.Product, error) {
	updated, err := scanProduct(r.db.QueryRowContext(ctx, queryUpdate,
		u.Photo, u.Title, u.Description, u.Price,
		u.DiscountPercentage, u.Quantity, u.Category, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Product{}, apperrors.NewNotFound("product not found")
	}
	if err != nil {
		return domain.Product{}, internal("update", err)
	}
	return updated, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id int, status string) error {
	result, err := r.db.ExecContext(ctx, queryUpdateStatus, status, id)
	if err != nil {
		return internal("update status", err)
	}
	return requireAffected(result, "update status")
}

func (r *repository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, queryDelete, id)
	if err != nil {
		return internal("delete", err)
	}
	return requireAffected(result, "delete")
}

// HasActiveOrders reports whether the product is in a pending, paid or shipped
// order. Before the order tables exist (module 05) no order can reference the
// product, so it returns false instead of failing.
func (r *repository) HasActiveOrders(ctx context.Context, productID int) (bool, error) {
	var tablesExist bool
	if err := r.db.QueryRowContext(ctx, queryOrderTablesExist).Scan(&tablesExist); err != nil {
		return false, internal("check order tables", err)
	}
	if !tablesExist {
		return false, nil
	}

	var exists bool
	if err := r.db.QueryRowContext(ctx, queryHasActiveOrders, productID).Scan(&exists); err != nil {
		return false, internal("has active orders", err)
	}
	return exists, nil
}

func requireAffected(result sql.Result, op string) error {
	n, err := result.RowsAffected()
	if err != nil {
		return internal(op, err)
	}
	if n == 0 {
		return apperrors.NewNotFound("product not found")
	}
	return nil
}
