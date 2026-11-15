package listproducts

import (
	"context"
	"strings"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) List(ctx context.Context, req ListRequest) (ListResponse, error) {
	if err := validate(req); err != nil {
		return ListResponse{}, err
	}

	limit, offset := paginate(req.Limit, req.Offset)

	products, total, err := u.repo.FindAll(ctx, shared.ProductFilters{
		Category: nonBlank(req.Category),
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Search:   nonBlank(req.Search),
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

func validate(req ListRequest) error {
	if req.MinPrice != nil && *req.MinPrice < 0 {
		return apperrors.NewBadRequest("min_price must be greater than or equal to 0")
	}
	if req.MaxPrice != nil && *req.MaxPrice < 0 {
		return apperrors.NewBadRequest("max_price must be greater than or equal to 0")
	}
	if req.MinPrice != nil && req.MaxPrice != nil && *req.MinPrice > *req.MaxPrice {
		return apperrors.NewBadRequest("min_price must not be greater than max_price")
	}
	return nil
}

// paginate applies the same rules as the admin user listing:
// an invalid limit falls back to the default and is capped at maxLimit.
func paginate(limit, offset int) (int, int) {
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func nonBlank(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
