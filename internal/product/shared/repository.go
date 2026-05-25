package shared

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type ProductFilters struct {
	Category *string
	MinPrice *float64
	MaxPrice *float64
	Search   *string
}

// ProductUpdate contém apenas campos opcionais (pointer) usados em partial update.
// Campos nil são ignorados pelo COALESCE no SQL — o valor existente é mantido.
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
	FindAll(filters ProductFilters, limit, offset int) ([]domain.Product, int, error)
	FindById(id int) (*domain.Product, error)
	Create(p domain.Product) (domain.Product, error)
	Update(id int, u ProductUpdate) (domain.Product, error)
	UpdateStatus(id int, status string) error
	Delete(id int) error
	HasActiveOrders(productID int) (bool, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(filters ProductFilters, limit, offset int) ([]domain.Product, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "status = 'active'")

	if filters.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *filters.Category)
		argIdx++
	}
	if filters.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argIdx))
		args = append(args, *filters.MinPrice)
		argIdx++
	}
	if filters.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argIdx))
		args = append(args, *filters.MaxPrice)
		argIdx++
	}
	if filters.Search != nil {
		conditions = append(conditions, fmt.Sprintf("LOWER(title) LIKE $%d", argIdx))
		args = append(args, "%"+strings.ToLower(*filters.Search)+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, apperrors.NewInternalServer(err.Error())
	}

	// Fetch page
	listArgs := append(args, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, seller_id, photo, title, description, price, discount_percentage,
		       quantity, status, category, created_at, updated_at
		FROM products %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, apperrors.NewInternalServer(err.Error())
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID, &p.SellerID, &p.Photo, &p.Title, &p.Description,
			&p.Price, &p.DiscountPercentage, &p.Quantity,
			&p.Status, &p.Category, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, apperrors.NewInternalServer(err.Error())
		}
		products = append(products, p)
	}

	return products, total, nil
}

func (r *repository) FindById(id int) (*domain.Product, error) {
	var p domain.Product
	err := r.db.QueryRow(queryFindById, id).Scan(
		&p.ID, &p.SellerID, &p.Photo, &p.Title, &p.Description,
		&p.Price, &p.DiscountPercentage, &p.Quantity,
		&p.Status, &p.Category, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewNotFound("product not found")
	}
	if err != nil {
		return nil, apperrors.NewInternalServer(err.Error())
	}
	return &p, nil
}

func (r *repository) Create(p domain.Product) (domain.Product, error) {
	err := r.db.QueryRow(
		queryCreate,
		p.SellerID, p.Photo, p.Title, p.Description,
		p.Price, p.DiscountPercentage, p.Quantity, p.Status, p.Category,
	).Scan(
		&p.ID, &p.SellerID, &p.Photo, &p.Title, &p.Description,
		&p.Price, &p.DiscountPercentage, &p.Quantity,
		&p.Status, &p.Category, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return domain.Product{}, apperrors.NewInternalServer(err.Error())
	}
	return p, nil
}

func (r *repository) Update(id int, u ProductUpdate) (domain.Product, error) {
	var p domain.Product
	err := r.db.QueryRow(
		queryUpdate,
		u.Photo, u.Title, u.Description, u.Price,
		u.DiscountPercentage, u.Quantity, u.Category, id,
	).Scan(
		&p.ID, &p.SellerID, &p.Photo, &p.Title, &p.Description,
		&p.Price, &p.DiscountPercentage, &p.Quantity,
		&p.Status, &p.Category, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Product{}, apperrors.NewNotFound("product not found")
	}
	if err != nil {
		return domain.Product{}, apperrors.NewInternalServer(err.Error())
	}
	return p, nil
}

func (r *repository) UpdateStatus(id int, status string) error {
	result, err := r.db.Exec(queryUpdateStatus, status, id)
	if err != nil {
		return apperrors.NewInternalServer(err.Error())
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NewNotFound("product not found")
	}
	return nil
}

func (r *repository) Delete(id int) error {
	result, err := r.db.Exec(queryDelete, id)
	if err != nil {
		return apperrors.NewInternalServer(err.Error())
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return apperrors.NewNotFound("product not found")
	}
	return nil
}

func (r *repository) HasActiveOrders(productID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(queryHasActiveOrders, productID).Scan(&exists)
	if err != nil {
		return false, apperrors.NewInternalServer(err.Error())
	}
	return exists, nil
}
