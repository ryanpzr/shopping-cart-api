package updateproduct

import "github.com/ryanpzr/shopping-cart-api/internal/product/shared"

type Usecase interface {
	Update(productID, userID int, req UpdateRequest) (ProductResponse, error)
}

type Repository = shared.Repository
