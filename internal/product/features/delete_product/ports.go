package deleteproduct

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Usecase interface {
	Delete(ctx context.Context, productID, userID int, role string) error
}

type Repository = shared.Repository
