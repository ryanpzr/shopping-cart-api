package adminupdatestatus

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/order/shared"
)

type Usecase interface {
	UpdateStatus(ctx context.Context, orderID, actorID int, req UpdateStatusRequest) (OrderResponse, error)
}

type (
	UnitOfWork     = shared.UnitOfWork
	TxRepositories = shared.TxRepositories
)
