package adminmanageuser

import (
	"time"

	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) guardNotAdmin(targetID int) (*domain.User, error) {
	user, err := u.repo.FindByID(targetID)
	if err != nil {
		return nil, err
	}
	if user.Role == "admin" {
		return nil, apperrors.NewForbidden("cannot manage an admin user")
	}
	return user, nil
}

func mapToResponse(u domain.User) ManageUserResponse {
	return ManageUserResponse{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		Role:         u.Role,
		Status:       u.Status,
		TimeoutUntil: u.TimeoutUntil,
		UpdatedAt:    u.UpdatedAt,
	}
}

func (u *usecase) BanUser(targetID int) (ManageUserResponse, error) {
	if _, err := u.guardNotAdmin(targetID); err != nil {
		return ManageUserResponse{}, err
	}
	updated, err := u.repo.UpdateStatus(targetID, "banned", nil)
	if err != nil {
		return ManageUserResponse{}, err
	}
	return mapToResponse(updated), nil
}

func (u *usecase) TimeoutUser(targetID, durationHours int) (ManageUserResponse, error) {
	if _, err := u.guardNotAdmin(targetID); err != nil {
		return ManageUserResponse{}, err
	}
	until := time.Now().Add(time.Duration(durationHours) * time.Hour)
	updated, err := u.repo.UpdateStatus(targetID, "timeout", &until)
	if err != nil {
		return ManageUserResponse{}, err
	}
	return mapToResponse(updated), nil
}

func (u *usecase) UnbanUser(targetID int) (ManageUserResponse, error) {
	if _, err := u.guardNotAdmin(targetID); err != nil {
		return ManageUserResponse{}, err
	}
	updated, err := u.repo.UpdateStatus(targetID, "active", nil)
	if err != nil {
		return ManageUserResponse{}, err
	}
	return mapToResponse(updated), nil
}
