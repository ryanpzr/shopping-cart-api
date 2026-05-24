package adminmanageuser

import shared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"

type Usecase interface {
	BanUser(targetID int) (ManageUserResponse, error)
	TimeoutUser(targetID, durationHours int) (ManageUserResponse, error)
	UnbanUser(targetID int) (ManageUserResponse, error)
}

type Repository = shared.Repository
