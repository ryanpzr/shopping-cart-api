package listorders

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/order/shared"
)

type Usecase interface {
	List(ctx context.Context, buyerID int, req ListRequest) (ListResponse, error)
}

type Repository = shared.Repository
