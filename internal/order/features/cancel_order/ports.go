package cancelorder

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/order/shared"
)

type Usecase interface {
	Cancel(ctx context.Context, orderID, userID int) (OrderResponse, error)
}

type (
	UnitOfWork     = shared.UnitOfWork
	TxRepositories = shared.TxRepositories
)
