package togglestatus

import "github.com/ryanpzr/shopping-cart-api/internal/product/shared"

type Usecase interface {
	Toggle(productID, userID int) (ToggleResponse, error)
}

type Repository = shared.Repository
