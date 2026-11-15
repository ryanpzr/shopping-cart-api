package listproducts

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Usecase interface {
	List(ctx context.Context, req ListRequest) (ListResponse, error)
}

type Repository = shared.Repository
