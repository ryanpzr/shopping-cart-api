package listproducts

import "github.com/ryanpzr/shopping-cart-api/internal/product/shared"

type Usecase interface {
	List(req ListRequest) (ListResponse, error)
}

type Repository = shared.Repository
