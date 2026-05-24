package getproduct

import "github.com/ryanpzr/shopping-cart-api/internal/product/shared"

type Usecase interface {
	Get(id int) (ProductResponse, error)
}

type Repository = shared.Repository
