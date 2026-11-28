package deleteproduct

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

const roleAdmin = "admin"

type usecase struct {
	uow UnitOfWork
}

func NewUsecase(uow UnitOfWork) Usecase {
	return &usecase{uow: uow}
}

// Delete implements BR-PROD-006. Everything runs in one transaction with the
// product row locked: checkout also locks the product row (after the cart), so
// no order can be placed for the product between the check and the delete.
func (u *usecase) Delete(ctx context.Context, productID, userID int, role string) error {
	err := u.uow.Do(ctx, func(r TxRepositories) error {
		p, err := r.Products.LockByID(ctx, productID)
		if err != nil {
			return err
		}

		// Admins can delete any product; clients only their own.
		if role != roleAdmin && p.SellerID != userID {
			return apperrors.NewForbidden("you are not the owner of this product")
		}

		hasOrders, err := r.Orders.HasOrdersWithStatus(ctx, productID, domain.ActiveOrderStatuses)
		if err != nil {
			return err
		}
		if hasOrders {
			return apperrors.NewConflict("product has active orders and cannot be deleted")
		}

		if err := r.Products.Delete(ctx, productID); err != nil {
			return err
		}

		// The event belongs to whoever deleted the product (the owner or an
		// admin) and is committed with the delete, or not at all.
		return r.Outbox.Insert(ctx, outbox.New(userID, outbox.EventProductDeleted, map[string]any{
			"product_id": productID,
			"title":      p.Title,
			"seller_id":  p.SellerID,
		}))
	})
	if err != nil {
		return toAppError(err)
	}
	return nil
}

func toAppError(err error) error {
	var appErr *apperrors.AppError
	switch {
	case errors.As(err, &appErr):
		return appErr
	case errors.Is(err, domain.ErrProductNotFound):
		return apperrors.NewNotFound("product not found")
	case errors.Is(err, domain.ErrProductInUse):
		// An order item that is no longer active (delivered, cancelled) still
		// references the product. Policy: the product stays, the seller can
		// deactivate it instead (see 05-order.md).
		return apperrors.NewConflict("product has orders and cannot be deleted")
	default:
		return apperrors.NewInternalServer(err.Error())
	}
}
