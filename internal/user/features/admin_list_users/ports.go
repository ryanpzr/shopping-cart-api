package adminlistusers

import shared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"

type Usecase interface {
	ListUsers(page, limit int) (PaginatedResponse, error)
}

type Repository = shared.Repository
