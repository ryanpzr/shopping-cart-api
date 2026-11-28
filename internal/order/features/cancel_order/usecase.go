package cancelorder

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/order/cancellation"
	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	uow UnitOfWork
}

func NewUsecase(uow UnitOfWork) Usecase {
	return &usecase{uow: uow}
}

// Cancel implements BR-ORDER-003: the owner cancels a pending order and the
// stock of every item goes back to its product, all in one transaction. The
// order row is locked, so two cancellations of the same order cannot both
// give the stock back.
func (u *usecase) Cancel(ctx context.Context, orderID, userID int) (OrderResponse, error) {
	var result domain.Order
	err := u.uow.Do(ctx, func(r TxRepositories) error {
		order, err := r.Orders.LockByID(ctx, orderID)
		if err != nil {
			return err
		}
		if !order.OwnedBy(userID) {
			return apperrors.NewForbidden("you are not the owner of this order")
		}

		if err := cancellation.Cancel(ctx, r, order); err != nil {
			return err
		}

		// TODO(module-06): emit `order_cancelled` through the outbox, in this transaction.

		result, err = r.Orders.FindByID(ctx, orderID)
		return err
	})
	if err != nil {
		return OrderResponse{}, toAppError(err)
	}
	return toResponse(result), nil
}

func toAppError(err error) error {
	var (
		appErr     *apperrors.AppError
		transition *domain.TransitionError
	)
	switch {
	case errors.As(err, &appErr):
		return appErr
	case errors.As(err, &transition):
		return apperrors.NewConflict("only a pending order can be cancelled")
	case errors.Is(err, domain.ErrOrderNotFound):
		return apperrors.NewNotFound("order not found")
	default:
		return apperrors.NewInternalServer(err.Error())
	}
}
