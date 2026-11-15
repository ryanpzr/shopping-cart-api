package listproducts

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

const defaultLimit = 20

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) List(ctx context.Context, req ListRequest) (ListResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	products, total, err := u.repo.FindAll(ctx, shared.ProductFilters{
		Category: req.Category,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Search:   req.Search,
	}, limit, offset)
	if err != nil {
		return ListResponse{}, err
	}

	data := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		data = append(data, toResponse(p))
	}

	return ListResponse{
		Data: data,
		Meta: MetaResponse{Total: total, Limit: limit, Offset: offset},
	}, nil
}
