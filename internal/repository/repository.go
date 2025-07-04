package repository

import (
	"database/sql"
	"errors"
	"github.com/ryanpzr/shopping-cart-api/internal/model"
)

type Repository interface {
	GetAllProduct() ([]model.Product, error)
	PostNewProduct(product model.Product) (model.Product, error)
	ChangeInfoProduct(id int, product model.Product) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAllProduct() ([]model.Product, error) {
	rows, err := r.db.Query("SELECT * FROM product;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.Id, &p.Photo, &p.Title, &p.Description, &p.Price, &p.Quantity); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *repository) PostNewProduct(product model.Product) (model.Product, error) {
	query := "INSERT INTO product (photo, title, description, price, quantity) VALUES ($1, $2, $3, $4, $5);"
	_, err := r.db.Exec(query, product.Photo, product.Title, product.Description, product.Price, product.Quantity)
	if err != nil {
		return model.Product{}, err
	}

	return product, nil
}

func (r *repository) ChangeInfoProduct(idProduct int, product model.Product) error {
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
