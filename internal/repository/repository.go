package repository

import (
	"database/sql"

	"github.com/ryanpzr/shopping-cart-api/model"
)

type Repository interface {
	GetAllProduct() ([]model.Product, error)
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
