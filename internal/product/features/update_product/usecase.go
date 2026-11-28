package updateproduct

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
)

type usecase struct {
	uow UnitOfWork
}

func NewUsecase(uow UnitOfWork) Usecase {
	return &usecase{uow: uow}
}

// Update applies a partial update. It runs in one transaction with the product
// row locked, so two concurrent updates cannot overwrite each other's fields,
// and the audit event is committed with the change.
func (u *usecase) Update(ctx context.Context, productID, userID int, req UpdateRequest) (ProductResponse, error) {
	if req.isEmpty() {
		return ProductResponse{}, apperrors.NewBadRequest("at least one field must be provided")
	}

	var updated domain.Product
	err := u.uow.Do(ctx, func(r TxRepositories) error {
		existing, err := r.Products.LockByID(ctx, productID)
		if err != nil {
			return err
		}
		if existing.SellerID != userID {
			return apperrors.NewForbidden("you are not the owner of this product")
		}

		product, err := apply(*existing, req)
		if err != nil {
			return err
		}
		if err := product.Validate(); err != nil {
			return err
		}

		updated, err = r.Products.Update(ctx, product)
		if err != nil {
			return err
		}

		return r.Outbox.Insert(ctx, outbox.New(userID, outbox.EventProductUpdated, map[string]any{
			"product_id":     productID,
			"changed_fields": req.changedFields(),
		}))
	})
	if err != nil {
		return ProductResponse{}, toAppError(err)
	}

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
	var (
		appErr  *apperrors.AppError
		invalid *domain.ValidationError
	)
	switch {
	case errors.As(err, &appErr):
		return appErr
	case errors.As(err, &invalid):
		return apperrors.NewBadRequest(invalid.Message)
	case errors.Is(err, domain.ErrProductNotFound):
		return apperrors.NewNotFound("product not found")
	default:
		return apperrors.NewInternalServer(err.Error())
	}
}
