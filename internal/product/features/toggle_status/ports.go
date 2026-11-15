package togglestatus

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Usecase interface {
	Toggle(ctx context.Context, productID, userID int) (ToggleResponse, error)
}

type Repository = shared.Repository
