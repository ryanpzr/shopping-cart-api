package createproduct

import "github.com/ryanpzr/shopping-cart-api/internal/product/domain"

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Create(sellerID int, req CreateRequest) (ProductResponse, error) {
	discount := 0
	if req.DiscountPercentage != nil {
		discount = *req.DiscountPercentage
	}

	p := domain.Product{
		SellerID:           sellerID,
		Photo:              req.Photo,
		Title:              req.Title,
		Description:        req.Description,
		Price:              req.Price,
		DiscountPercentage: discount,
		Quantity:           req.Quantity,
		Status:             "active",
		Category:           req.Category,
	}

	created, err := u.repo.Create(p)
	if err != nil {
		return ProductResponse{}, err
	}

	// TODO(module-06): Emitir evento 'product_created' no activity log.
	// Ver: sdd/specs/modules/06-activity-log.md

	finalPrice := created.Price * (1 - float64(created.DiscountPercentage)/100.0)
	return ProductResponse{
		ID:                 created.ID,
		SellerID:           created.SellerID,
		Photo:              created.Photo,
		Title:              created.Title,
		Description:        created.Description,
		Price:              created.Price,
		DiscountPercentage: created.DiscountPercentage,
		FinalPrice:         finalPrice,
		Quantity:           created.Quantity,
		Status:             created.Status,
		Category:           created.Category,
		CreatedAt:          created.CreatedAt,
		UpdatedAt:          created.UpdatedAt,
	}, nil
}
