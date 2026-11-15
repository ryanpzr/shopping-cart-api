package deleteproduct

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

const roleAdmin = "admin"

type usecase struct {
	repo   Repository
	orders OrderChecker
}

func NewUsecase(repo Repository, orders OrderChecker) Usecase {
	return &usecase{repo: repo, orders: orders}
}

func (u *usecase) Delete(ctx context.Context, productID, userID int, role string) error {
	p, err := u.repo.FindByID(ctx, productID)
	if err != nil {
		return toAppError(err)
	}

	// Admins can delete any product; clients only their own (BR-PROD-006).
	if role != roleAdmin && p.SellerID != userID {
		return apperrors.NewForbidden("you are not the owner of this product")
	}

	hasOrders, err := u.orders.HasOrdersWithStatus(ctx, productID, domain.ActiveOrderStatuses)
	if err != nil {
		return toAppError(err)
	}
	if hasOrders {
		return apperrors.NewConflict("product has active orders and cannot be deleted")
	}

	if err := u.repo.Delete(ctx, productID); err != nil {
		return toAppError(err)
	}

	// TODO(module-06): emit `product_deleted` through the outbox, in the same transaction.

	return nil
}

func toAppError(err error) error {
	if errors.Is(err, domain.ErrProductNotFound) {
		return apperrors.NewNotFound("product not found")
	}
	return apperrors.NewInternalServer(err.Error())
}
