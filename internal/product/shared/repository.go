package shared

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/database"
)

// DBTX is satisfied by both *sql.DB and *sql.Tx, so the repository can run
// inside a transaction owned by the caller (e.g. checkout, outbox events).
type DBTX = database.DBTX

// ProductFilters narrows FindAll. Nil fields are not applied; the repository
// adds no implicit condition, so visibility rules belong to the caller.
type ProductFilters struct {
	Status   *string
	Category *string
	MinPrice *float64
	MaxPrice *float64
	Search   *string
}

// Repository only talks to the database. It returns domain.ErrProductNotFound
// when no product matches and wraps every other error with context; mapping
// errors to application responses is the usecase's job.
type Repository interface {
	FindAll(ctx context.Context, filters ProductFilters, limit, offset int) ([]domain.Product, int, error)
	FindByID(ctx context.Context, id int) (*domain.Product, error)
	Create(ctx context.Context, p domain.Product) (domain.Product, error)
	Update(ctx context.Context, p domain.Product) (domain.Product, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	Delete(ctx context.Context, id int) error
	// LockByIDs loads the products with SELECT … FOR UPDATE. It must run
	// inside a transaction; missing ids are simply absent from the result.
	LockByIDs(ctx context.Context, ids []int) ([]domain.Product, error)
	UpdateStock(ctx context.Context, id, quantity int) error
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

// PostgreSQL error code for foreign_key_violation.
const pgForeignKeyViolation = "23503"

func wrap(op string, err error) error {
	return fmt.Errorf("product repository: %s: %w", op, err)
}

func (r *repository) FindAll(ctx context.Context, filters ProductFilters, limit, offset int) ([]domain.Product, int, error) {
	var conditions []string
	var args []any

	addCondition := func(format string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(format, len(args)))
	}
	if filters.Status != nil {
		addCondition("status = $%d", *filters.Status)
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
		addCondition(`title ILIKE $%d ESCAPE '\'`, "%"+escapeLike(*filters.Search)+"%")
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM products " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, wrap("count", err)
	}

	listQuery := fmt.Sprintf(`
		SELECT %s
		FROM products %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, productColumns, where, len(args)+1, len(args)+2)

	rows, err := r.db.QueryContext(ctx, listQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, wrap("list", err)
	}
	defer func() { _ = rows.Close() }()

	products := make([]domain.Product, 0, limit)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, wrap("scan", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, wrap("iterate", err)
	}

	return products, total, nil
}

func (r *repository) FindByID(ctx context.Context, id int) (*domain.Product, error) {
	p, err := scanProduct(r.db.QueryRowContext(ctx, queryFindByID, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, wrap("find by id", err)
	}
	return &p, nil
}

func (r *repository) Create(ctx context.Context, p domain.Product) (domain.Product, error) {
	created, err := scanProduct(r.db.QueryRowContext(ctx, queryCreate,
		p.SellerID, p.Photo, p.Title, p.Description,
		p.Price, p.DiscountPercentage, p.Quantity, p.Status, p.Category,
	))
	if err != nil {
		return domain.Product{}, wrap("create", err)
	}
	return created, nil
}

// Update writes every editable column of p. The caller loads the product,
// applies the changes and validates the result before calling it.
func (r *repository) Update(ctx context.Context, p domain.Product) (domain.Product, error) {
	updated, err := scanProduct(r.db.QueryRowContext(ctx, queryUpdate,
		p.Photo, p.Title, p.Description, p.Price,
		p.DiscountPercentage, p.Quantity, p.Category, p.ID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Product{}, domain.ErrProductNotFound
	}
	if err != nil {
		return domain.Product{}, wrap("update", err)
	}
	return updated, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id int, status string) error {
	result, err := r.db.ExecContext(ctx, queryUpdateStatus, status, id)
	if err != nil {
		return wrap("update status", err)
	}
	return requireAffected(result, "update status")
}

func (r *repository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, queryDelete, id)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == pgForeignKeyViolation {
		return domain.ErrProductInUse
	}
	if err != nil {
		return wrap("delete", err)
	}
	return requireAffected(result, "delete")
}

func (r *repository) LockByIDs(ctx context.Context, ids []int) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(ctx, queryLockByIDs, pq.Array(ids))
	if err != nil {
		return nil, wrap("lock by ids", err)
	}
	defer func() { _ = rows.Close() }()

	products := make([]domain.Product, 0, len(ids))
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, wrap("scan", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, wrap("iterate", err)
	}
	return products, nil
}

func (r *repository) UpdateStock(ctx context.Context, id, quantity int) error {
	result, err := r.db.ExecContext(ctx, queryUpdateStock, quantity, id)
	if err != nil {
		return wrap("update stock", err)
	}
	return requireAffected(result, "update stock")
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// escapeLike makes LIKE wildcards in user input match literally, so a search
// for "50%" does not match every title containing "50".
func escapeLike(s string) string {
	return likeEscaper.Replace(s)
}

func requireAffected(result sql.Result, op string) error {
	n, err := result.RowsAffected()
	if err != nil {
		return wrap(op, err)
	}
	if n == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}
