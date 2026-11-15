package updateproduct

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Update(ctx context.Context, productID, userID int, req UpdateRequest) (ProductResponse, error) {
	if err := validate(req); err != nil {
		return ProductResponse{}, err
	}

	existing, err := u.repo.FindByID(ctx, productID)
	if err != nil {
		return ProductResponse{}, err
	}
	if existing.SellerID != userID {
		return ProductResponse{}, apperrors.NewForbidden("you are not the owner of this product")
	}

	updated, err := u.repo.Update(ctx, productID, shared.ProductUpdate{
		Photo:              req.Photo,
		Title:              req.Title,
		Description:        req.Description,
		Price:              req.Price,
		DiscountPercentage: req.DiscountPercentage,
		Quantity:           req.Quantity,
		Category:           req.Category,
	})
	if err != nil {
		return ProductResponse{}, err
	}

	// TODO(module-06): emit `product_updated` through the outbox, in the same transaction.

	return toResponse(updated), nil
}

func validate(req UpdateRequest) error {
	if req.isEmpty() {
		return apperrors.NewBadRequest("at least one field must be provided")
	}
	if req.Title != nil {
		if err := domain.ValidateTitle(*req.Title); err != nil {
			return err
		}
	}
	if req.Price != nil {
		if err := domain.ValidatePrice(*req.Price); err != nil {
			return err
		}
	}
	if req.DiscountPercentage != nil {
		if err := domain.ValidateDiscount(*req.DiscountPercentage); err != nil {
			return err
		}
	}
	if req.Quantity != nil {
		if err := domain.ValidateQuantity(*req.Quantity); err != nil {
			return err
		}
	}
	return domain.ValidateCategory(req.Category)
}
