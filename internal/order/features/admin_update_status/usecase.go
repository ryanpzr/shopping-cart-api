package adminupdatestatus

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

// UpdateStatus implements BR-ADMIN-ORDER-001: the admin moves an order to its
// next status (pending → paid → shipped → delivered) or cancels a pending
// order, which gives the stock back like the buyer's cancellation. Any other
// change is a conflict. The order row is locked, so concurrent changes run
// one after the other.
func (u *usecase) UpdateStatus(ctx context.Context, orderID int, req UpdateStatusRequest) (OrderResponse, error) {
	if req.Status == nil {
		return OrderResponse{}, apperrors.NewBadRequest("status is required")
	}
	if err := domain.ValidateStatus(*req.Status); err != nil {
		return OrderResponse{}, toAppError(err)
	}
	target := *req.Status

	var result domain.Order
	err := u.uow.Do(ctx, func(r TxRepositories) error {
		order, err := r.Orders.LockByID(ctx, orderID)
		if err != nil {
			return err
		}

		if target == domain.StatusCancelled {
			err = cancellation.Cancel(ctx, r, order)
		} else if err = domain.CheckTransition(order.Status, target); err == nil {
			err = r.Orders.UpdateStatus(ctx, orderID, target)
		}
		if err != nil {
			return err
		}

		// TODO(module-06): emit `order_cancelled` through the outbox when target is cancelled.

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
		invalid    *domain.ValidationError
		transition *domain.TransitionError
	)
	switch {
	case errors.As(err, &invalid):
		return apperrors.NewBadRequest(invalid.Message)
	case errors.As(err, &transition):
		return apperrors.NewConflict(transition.Error())
	case errors.Is(err, domain.ErrOrderNotFound):
		return apperrors.NewNotFound("order not found")
	default:
		return apperrors.NewInternalServer(err.Error())
	}
}
