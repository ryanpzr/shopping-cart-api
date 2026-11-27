package removeitem

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	uow UnitOfWork
}

func NewUsecase(uow UnitOfWork) Usecase {
	return &usecase{uow: uow}
}

// RemoveItem implements BR-CART-003: the item must belong to the buyer's
// active cart. The cart row is locked, so a concurrent checkout cannot close
// the cart between the lookup and the delete.
func (u *usecase) RemoveItem(ctx context.Context, buyerID, itemID int) (CartResponse, error) {
	var cart domain.Cart
	err := u.uow.Do(ctx, func(r TxRepositories) error {
		locked, err := r.Carts.LockByBuyer(ctx, buyerID, domain.StatusActive)
		if err != nil {
			return err
		}
		if err := r.Carts.RemoveItem(ctx, locked.ID, itemID); err != nil {
			return err
		}
		// Read the result in the same transaction, while the cart is still locked.
		cart, err = r.Carts.FindByBuyer(ctx, buyerID, domain.StatusActive)
		return err
	})
	if err != nil {
		return CartResponse{}, toAppError(err)
	}
	return toResponse(cart), nil
}

func toAppError(err error) error {
	if errors.Is(err, domain.ErrItemNotFound) || errors.Is(err, domain.ErrCartNotFound) {
		return apperrors.NewNotFound("cart item not found")
	}
	return apperrors.NewInternalServer(err.Error())
}
