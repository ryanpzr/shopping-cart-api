package createproduct

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

func (u *usecase) Create(ctx context.Context, sellerID int, req CreateRequest) (ProductResponse, error) {
	if req.Price == nil {
		return ProductResponse{}, apperrors.NewBadRequest("price is required")
	}
	if req.Quantity == nil {
		return ProductResponse{}, apperrors.NewBadRequest("quantity is required")
	}

	discount := 0
	if req.DiscountPercentage != nil {
		discount = *req.DiscountPercentage
	}

	product := domain.Product{
		SellerID:           sellerID,
		Photo:              req.Photo,
		Title:              req.Title,
		Description:        req.Description,
		Price:              *req.Price,
		DiscountPercentage: discount,
		Quantity:           *req.Quantity,
		Status:             domain.StatusActive,
		Category:           req.Category,
	}
	if err := product.Validate(); err != nil {
		return ProductResponse{}, toAppError(err)
	}

	created, err := u.repo.Create(ctx, product)
	if err != nil {
		return ProductResponse{}, toAppError(err)
	}

	// TODO(module-06): emit `product_created` through the outbox, in the same transaction.

	return toResponse(created), nil
}

func toAppError(err error) error {
	var invalid *domain.ValidationError
	if errors.As(err, &invalid) {
		return apperrors.NewBadRequest(invalid.Message)
	}
	return apperrors.NewInternalServer(err.Error())
}
