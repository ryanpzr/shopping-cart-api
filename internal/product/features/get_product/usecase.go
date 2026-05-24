package getproduct

import "github.com/ryanpzr/shopping-cart-api/pkg/apperrors"

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Get(id int) (ProductResponse, error) {
	p, err := u.repo.FindById(id)
	if err != nil {
		return ProductResponse{}, err
	}

	if p.Status == "inactive" {
		return ProductResponse{}, apperrors.NewNotFound("product not found")
	}

	finalPrice := p.Price * (1 - float64(p.DiscountPercentage)/100.0)
	return ProductResponse{
		ID:                 p.ID,
		SellerID:           p.SellerID,
		Photo:              p.Photo,
		Title:              p.Title,
		Description:        p.Description,
		Price:              p.Price,
		DiscountPercentage: p.DiscountPercentage,
		FinalPrice:         finalPrice,
		Quantity:           p.Quantity,
		Status:             p.Status,
		Category:           p.Category,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}, nil
}
