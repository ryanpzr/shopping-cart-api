package updateproduct

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

func (u *usecase) Update(ctx context.Context, productID, userID int, req UpdateRequest) (ProductResponse, error) {
	if req.isEmpty() {
		return ProductResponse{}, apperrors.NewBadRequest("at least one field must be provided")
	}

	existing, err := u.repo.FindByID(ctx, productID)
	if err != nil {
		return ProductResponse{}, toAppError(err)
	}
	if existing.SellerID != userID {
		return ProductResponse{}, apperrors.NewForbidden("you are not the owner of this product")
	}

	product, err := apply(*existing, req)
	if err != nil {
		return ProductResponse{}, err
	}
	if err := product.Validate(); err != nil {
		return ProductResponse{}, toAppError(err)
	}

	updated, err := u.repo.Update(ctx, product)
	if err != nil {
		return ProductResponse{}, toAppError(err)
	}

	// TODO(module-06): emit `product_updated` through the outbox, in the same transaction.

	return toResponse(updated), nil
}

// apply returns p with the fields present in req. Nullable fields are cleared
// by null; required fields reject it.
func apply(p domain.Product, req UpdateRequest) (domain.Product, error) {
	if req.Photo.Set {
		p.Photo = req.Photo.Value
	}
	if req.Description.Set {
		p.Description = req.Description.Value
	}
	if req.Category.Set {
		p.Category = req.Category.Value
	}
	if req.Title.Set {
		if req.Title.Value == nil {
			return p, notNull("title")
		}
		p.Title = *req.Title.Value
	}
	if req.Price.Set {
		if req.Price.Value == nil {
			return p, notNull("price")
		}
		p.Price = *req.Price.Value
	}
	if req.DiscountPercentage.Set {
		if req.DiscountPercentage.Value == nil {
			return p, notNull("discount_percentage")
		}
		p.DiscountPercentage = *req.DiscountPercentage.Value
	}
	if req.Quantity.Set {
		if req.Quantity.Value == nil {
			return p, notNull("quantity")
		}
		p.Quantity = *req.Quantity.Value
	}
	return p, nil
}

func notNull(field string) error {
	return apperrors.NewBadRequest(field + " cannot be null")
}

func toAppError(err error) error {
	var invalid *domain.ValidationError
	if errors.As(err, &invalid) {
		return apperrors.NewBadRequest(invalid.Message)
	}
	if errors.Is(err, domain.ErrProductNotFound) {
		return apperrors.NewNotFound("product not found")
	}
	return apperrors.NewInternalServer(err.Error())
}
