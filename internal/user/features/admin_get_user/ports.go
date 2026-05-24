package admingetuser

import shared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"

type Usecase interface {
	GetUser(id int) (AdminUserResponse, error)
}

type Repository = shared.Repository
