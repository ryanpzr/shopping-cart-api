package deleteproduct

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

const roleAdmin = "admin"

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Delete(ctx context.Context, productID, userID int, role string) error {
	p, err := u.repo.FindByID(ctx, productID)
	if err != nil {
		return err
	}

	// Admins can delete any product; clients only their own (BR-PROD-006).
	if role != roleAdmin && p.SellerID != userID {
		return apperrors.NewForbidden("you are not the owner of this product")
	}

	hasOrders, err := u.repo.HasActiveOrders(ctx, productID)
	if err != nil {
		return err
	}
	if hasOrders {
		return apperrors.NewConflict("product has active orders and cannot be deleted")
	}

	if err := u.repo.Delete(ctx, productID); err != nil {
		return err
	}

	// TODO(module-06): emit `product_deleted` through the outbox, in the same transaction.

	return nil
}
