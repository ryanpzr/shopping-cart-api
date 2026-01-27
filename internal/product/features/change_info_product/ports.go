package changeinfoproduct

import (
	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Usecase interface {
	ChangeInfoProduct(id int, product domain.Product) error
	GetAllProduct() ([]domain.Product, error)
	PostNewProduct(product domain.Product) (domain.Product, error)
}

type Repository = shared.Repository
