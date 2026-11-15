package getproduct

import (
	"context"

	"github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

type Usecase interface {
	Get(ctx context.Context, id int) (ProductResponse, error)
}

type Repository = shared.Repository
