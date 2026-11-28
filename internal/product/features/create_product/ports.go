package createproduct

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Usecase interface {
	Create(ctx context.Context, sellerID int, req CreateRequest) (ProductResponse, error)
}

type (
	UnitOfWork     = shared.UnitOfWork
	TxRepositories = shared.TxRepositories
)
