package updateme

import shared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"

type Usecase interface {
	UpdateMe(userID int, req UpdateMeRequest) (UpdateMeResponse, error)
}

type Repository = shared.Repository
