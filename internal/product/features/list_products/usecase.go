package listproducts

import (
	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) List(req ListRequest) (ListResponse, error) {
	filters := shared.ProductFilters{
		Category: req.Category,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Search:   req.Search,
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	products, total, err := u.repo.FindAll(filters, limit, offset)
	if err != nil {
		return ListResponse{}, err
	}

	data := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		data = append(data, toProductResponse(p))
	}

	return ListResponse{
		Data: data,
		Meta: MetaResponse{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

func toProductResponse(p domain.Product) ProductResponse {
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
	}
}
