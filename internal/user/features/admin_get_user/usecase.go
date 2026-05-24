package admingetuser

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) GetUser(id int) (AdminUserResponse, error) {
	user, err := u.repo.FindByID(id)
	if err != nil {
		return AdminUserResponse{}, err
	}
	return AdminUserResponse{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		Status:       user.Status,
		TimeoutUntil: user.TimeoutUntil,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}
