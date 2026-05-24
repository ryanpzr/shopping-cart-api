package updateme

import "github.com/ryanpzr/shopping-cart-api/pkg/apperrors"

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) UpdateMe(userID int, req UpdateMeRequest) (UpdateMeResponse, error) {
	current, err := u.repo.FindByID(userID)
	if err != nil {
		return UpdateMeResponse{}, err
	}

	// Partial update: keep existing values when the request field is empty.
	// Role and status are not in the request body — protection is structural.
	finalName := current.Name
	if req.Name != "" {
		finalName = req.Name
	}

	finalEmail := current.Email
	if req.Email != "" && req.Email != current.Email {
		existing, err := u.repo.FindByEmail(req.Email)
		if err == nil && existing != nil {
			return UpdateMeResponse{}, apperrors.NewConflict("email already in use")
		}
		finalEmail = req.Email
	}

	updated, err := u.repo.UpdateProfile(userID, finalName, finalEmail)
	if err != nil {
		return UpdateMeResponse{}, err
	}

	return UpdateMeResponse{
		ID:        updated.ID,
		Name:      updated.Name,
		Email:     updated.Email,
		Role:      updated.Role,
		Status:    updated.Status,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}
