package removeitem

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

// RemoveItem implements BR-CART-003: the item must belong to the buyer's active cart.
func (u *usecase) RemoveItem(ctx context.Context, buyerID, itemID int) (CartResponse, error) {
	cart, err := u.repo.FindByBuyer(ctx, buyerID, domain.StatusActive)
	if err != nil {
		return CartResponse{}, toAppError(err)
	}
	if err := u.repo.RemoveItem(ctx, cart.ID, itemID); err != nil {
		return CartResponse{}, toAppError(err)
	}

	cart, err = u.repo.FindByBuyer(ctx, buyerID, domain.StatusActive)
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
