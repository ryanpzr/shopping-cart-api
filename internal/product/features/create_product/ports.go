package createproduct

import "github.com/ryanpzr/shopping-cart-api/internal/product/shared"

type Usecase interface {
	Create(sellerID int, req CreateRequest) (ProductResponse, error)
}

type Repository = shared.Repository
