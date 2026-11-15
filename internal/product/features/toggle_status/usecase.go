package togglestatus

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Toggle(ctx context.Context, productID, userID int) (ToggleResponse, error) {
	p, err := u.repo.FindByID(ctx, productID)
	if err != nil {
		return ToggleResponse{}, err
	}
	if p.SellerID != userID {
		return ToggleResponse{}, apperrors.NewForbidden("you are not the owner of this product")
	}

	newStatus := domain.StatusInactive
	if !p.IsActive() {
		newStatus = domain.StatusActive
	}

	if err := u.repo.UpdateStatus(ctx, productID, newStatus); err != nil {
		return ToggleResponse{}, err
	}

	return ToggleResponse{ID: productID, Status: newStatus}, nil
}
