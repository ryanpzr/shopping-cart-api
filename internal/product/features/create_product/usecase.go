package createproduct

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

func (u *usecase) Create(ctx context.Context, sellerID int, req CreateRequest) (ProductResponse, error) {
	if err := validate(req); err != nil {
		return ProductResponse{}, err
	}

	discount := 0
	if req.DiscountPercentage != nil {
		discount = *req.DiscountPercentage
	}

	created, err := u.repo.Create(ctx, domain.Product{
		SellerID:           sellerID,
		Photo:              req.Photo,
		Title:              req.Title,
		Description:        req.Description,
		Price:              *req.Price,
		DiscountPercentage: discount,
		Quantity:           *req.Quantity,
		Status:             domain.StatusActive,
		Category:           req.Category,
	})
	if err != nil {
		return ProductResponse{}, err
	}

	// TODO(module-06): emit `product_created` through the outbox, in the same transaction.

	return toResponse(created), nil
}

func validate(req CreateRequest) error {
	if err := domain.ValidateTitle(req.Title); err != nil {
		return err
	}
	if req.Price == nil {
		return apperrors.NewBadRequest("price is required")
	}
	if err := domain.ValidatePrice(*req.Price); err != nil {
		return err
	}
	if req.Quantity == nil {
		return apperrors.NewBadRequest("quantity is required")
	}
	if err := domain.ValidateQuantity(*req.Quantity); err != nil {
		return err
	}
	if req.DiscountPercentage != nil {
		if err := domain.ValidateDiscount(*req.DiscountPercentage); err != nil {
			return err
		}
	}
	return domain.ValidateCategory(req.Category)
}
