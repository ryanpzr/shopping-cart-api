package removeitem

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared"
)

type Usecase interface {
	RemoveItem(ctx context.Context, buyerID, itemID int) (CartResponse, error)
}

type Repository = shared.Repository
