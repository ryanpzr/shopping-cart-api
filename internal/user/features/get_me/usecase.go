package getme

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) GetMe(userID int) (GetMeResponse, error) {
	user, err := u.repo.FindByID(userID)
	if err != nil {
		return GetMeResponse{}, err
	}
	return GetMeResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}, nil
}
