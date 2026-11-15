package getproduct

import (
	"context"

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
		return ProductResponse{}, err
	}

	// Inactive products are hidden from the public catalog (BR-PROD-002).
	if !p.IsActive() {
		return ProductResponse{}, apperrors.NewNotFound("product not found")
	}

	return toResponse(*p), nil
}
