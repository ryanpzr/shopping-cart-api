package updateproduct

import (
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Update(productID, userID int, req UpdateRequest) (ProductResponse, error) {
	existing, err := u.repo.FindById(productID)
	if err != nil {
		return ProductResponse{}, err
	}

	if existing.SellerID != userID {
		return ProductResponse{}, apperrors.NewForbidden("you are not the owner of this product")
	}

	updated, err := u.repo.Update(productID, shared.ProductUpdate{
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

	// TODO(module-06): Emitir evento 'product_updated' no activity log.
	// Ver: sdd/specs/modules/06-activity-log.md

	finalPrice := updated.Price * (1 - float64(updated.DiscountPercentage)/100.0)
	return ProductResponse{
		ID:                 updated.ID,
		SellerID:           updated.SellerID,
		Photo:              updated.Photo,
		Title:              updated.Title,
		Description:        updated.Description,
		Price:              updated.Price,
		DiscountPercentage: updated.DiscountPercentage,
		FinalPrice:         finalPrice,
		Quantity:           updated.Quantity,
		Status:             updated.Status,
		Category:           updated.Category,
		CreatedAt:          updated.CreatedAt,
		UpdatedAt:          updated.UpdatedAt,
	}, nil
}
