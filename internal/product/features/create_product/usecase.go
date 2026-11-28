package createproduct

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

	var created domain.Product
	err := u.uow.Do(ctx, func(r TxRepositories) error {
		var err error
		created, err = r.Products.Create(ctx, product)
		if err != nil {
			return err
		}
		// The audit event is committed with the product, or not at all.
		return r.Outbox.Insert(ctx, outbox.New(sellerID, outbox.EventProductCreated, map[string]any{
			"product_id": created.ID,
			"title":      created.Title,
		}))
	})
	if err != nil {
		return ProductResponse{}, toAppError(err)
	}

	return toResponse(created), nil
}

func toAppError(err error) error {
	var invalid *domain.ValidationError
	if errors.As(err, &invalid) {
		return apperrors.NewBadRequest(invalid.Message)
	}
	return apperrors.NewInternalServer(err.Error())
}
