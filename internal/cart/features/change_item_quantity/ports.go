package changeitemquantity

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared"
)

type Usecase interface {
	ChangeQuantity(ctx context.Context, buyerID, itemID int, req ChangeQuantityRequest) (CartResponse, error)
}

type (
	UnitOfWork     = shared.UnitOfWork
	TxRepositories = shared.TxRepositories
)
