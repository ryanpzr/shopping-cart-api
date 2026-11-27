package getorder

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/order/shared"
)

type Usecase interface {
	Get(ctx context.Context, orderID, userID int, role string) (OrderResponse, error)
}

type Repository = shared.Repository
