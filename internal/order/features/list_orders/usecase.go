package listorders

import (
	"context"

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

// List implements BR-ORDER-001: the buyer's own orders, newest first, with items.
func (u *usecase) List(ctx context.Context, buyerID int, req ListRequest) (ListResponse, error) {
	limit, offset := paginate(req.Limit, req.Offset)

	orders, total, err := u.repo.FindByBuyer(ctx, buyerID, limit, offset)
	if err != nil {
		return ListResponse{}, apperrors.NewInternalServer(err.Error())
	}

	data := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		data = append(data, toResponse(o))
	}
	return ListResponse{
		Data: data,
		Meta: MetaResponse{Total: total, Limit: limit, Offset: offset},
	}, nil
}

// paginate applies the same rules as the other listings: an invalid limit
// falls back to the default and is capped at maxLimit.
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
