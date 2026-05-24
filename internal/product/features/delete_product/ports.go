package deleteproduct

import "github.com/ryanpzr/shopping-cart-api/internal/product/shared"

type Usecase interface {
	Delete(productID, userID int, role string) error
}

type Repository = shared.Repository
