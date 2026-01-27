package shared

import (
	"database/sql"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
)

type Repository interface {
	GetAllProduct() ([]domain.Product, error)
	PostNewProduct(product domain.Product) (domain.Product, error)
	ChangeInfoProduct(id int, product domain.Product) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAllProduct() ([]domain.Product, error) {
	rows, err := r.db.Query("SELECT * FROM product;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.Id, &p.Photo, &p.Title, &p.Description, &p.Price, &p.Quantity); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *repository) PostNewProduct(product domain.Product) (domain.Product, error) {
	query := "INSERT INTO product (photo, title, description, price, quantity) VALUES ($1, $2, $3, $4, $5);"
	_, err := r.db.Exec(query, product.Photo, product.Title, product.Description, product.Price, product.Quantity)
	if err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

func (r *repository) ChangeInfoProduct(idProduct int, product domain.Product) error {
	query := `UPDATE product SET photo=$1, title=$2, description=$3, price=$4, quantity=$5 WHERE id=$6`
	result, err := r.db.Exec(query, product.Photo, product.Title, product.Description, product.Price, product.Quantity, idProduct)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}
