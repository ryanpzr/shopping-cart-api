package repository

import (
	"database/sql"

	"github.com/ryanpzr/shopping-cart-api/internal/model"
)

type Repository interface {
	GetAllProduct() ([]model.Product, error)
	Post(product model.Product) (model.Product, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAllProduct() ([]model.Product, error) {
	rows, err := r.db.Query("SELECT * FROM produto")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.Id, &p.Foto, &p.Titulo, &p.Descricao, &p.Preco, &p.Quantidade); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *repository) Post(product model.Product) (model.Product, error) {
	query := "INSERT INTO produto (foto, titulo, descricao, preco, quantidade) VALUES ($1, $2, $3, $4, $5)"
	_, err := r.db.Exec(query, product.Foto, product.Titulo, product.Descricao, product.Preco, product.Quantidade)
	if err != nil {
		return model.Product{}, err
	}

	return product, nil
}
