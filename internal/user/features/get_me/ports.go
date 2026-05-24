package getme

import shared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"

type Usecase interface {
	GetMe(userID int) (GetMeResponse, error)
}

type Repository = shared.Repository
