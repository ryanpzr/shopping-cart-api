package getproduct

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Get(ctx context.Context, id int) (ProductResponse, error) {
	p, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return ProductResponse{}, toAppError(err)
	}

	// Inactive products are hidden from the public catalog (BR-PROD-002).
	if !p.IsActive() {
		return ProductResponse{}, apperrors.NewNotFound("product not found")
	}

	return toResponse(*p), nil
}

func toAppError(err error) error {
	if errors.Is(err, domain.ErrProductNotFound) {
		return apperrors.NewNotFound("product not found")
	}
	return apperrors.NewInternalServer(err.Error())
}
