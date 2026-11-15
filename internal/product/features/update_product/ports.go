package updateproduct

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Usecase interface {
	Update(ctx context.Context, productID, userID int, req UpdateRequest) (ProductResponse, error)
}

type Repository = shared.Repository
