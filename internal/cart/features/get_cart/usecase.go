package getcart

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

// Get returns the buyer's active cart, creating it when there is none (BR-CART-001).
func (u *usecase) Get(ctx context.Context, buyerID int) (CartResponse, error) {
	cart, err := u.repo.FindByBuyer(ctx, buyerID, domain.StatusActive)
	if errors.Is(err, domain.ErrCartNotFound) {
		cart, err = u.repo.Create(ctx, buyerID, domain.StatusActive)
		if errors.Is(err, domain.ErrActiveCartExists) {
			// Another request created it first: read that one.
			cart, err = u.repo.FindByBuyer(ctx, buyerID, domain.StatusActive)
		}
	}
	if err != nil {
		return CartResponse{}, apperrors.NewInternalServer(err.Error())
	}
	return toResponse(cart), nil
}
