package additemtocart

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared"
)

type Usecase interface {
	AddItem(ctx context.Context, buyerID int, req AddItemRequest) (CartResponse, error)
}

type (
	Repository     = shared.Repository
	UnitOfWork     = shared.UnitOfWork
	TxRepositories = shared.TxRepositories
)
