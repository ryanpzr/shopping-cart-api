package adminlistusers

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) ListUsers(page, limit int) (PaginatedResponse, error) {
	offset := (page - 1) * limit

	users, total, err := u.repo.FindAll(limit, offset)
	if err != nil {
		return PaginatedResponse{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = total / limit
		if total%limit != 0 {
			totalPages++
		}
	}

	data := make([]UserSummary, len(users))
	for i, usr := range users {
		data[i] = UserSummary{
			ID:        usr.ID,
			Name:      usr.Name,
			Email:     usr.Email,
			Role:      usr.Role,
			Status:    usr.Status,
			CreatedAt: usr.CreatedAt,
		}
	}

	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
