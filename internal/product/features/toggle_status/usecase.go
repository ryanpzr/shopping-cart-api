package togglestatus

import "github.com/ryanpzr/shopping-cart-api/pkg/apperrors"

const (
	inactive = "inactive"
	active   = "active"
)

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Toggle(productID, userID int) (ToggleResponse, error) {
	p, err := u.repo.FindById(productID)
	if err != nil {
		return ToggleResponse{}, err
	}

	if p.SellerID != userID {
		return ToggleResponse{}, apperrors.NewForbidden("you are not the owner of this product")
	}

	newStatus := inactive
	if p.Status == inactive {
		newStatus = active
	}

	if err := u.repo.UpdateStatus(productID, newStatus); err != nil {
		return ToggleResponse{}, err
	}

	return ToggleResponse{ID: productID, Status: newStatus}, nil
}
