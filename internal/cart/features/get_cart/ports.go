package getcart

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared"
)

type Usecase interface {
	Get(ctx context.Context, buyerID int) (CartResponse, error)
}

type Repository = shared.Repository
