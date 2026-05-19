package register

import (
	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	usershared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	jwtutil "github.com/ryanpzr/shopping-cart-api/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type usecase struct {
	repo usershared.Repository
}

func NewUsecase(repo usershared.Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Register(req RegisterRequest) (RegisterResponse, error) {
	existing, _ := u.repo.FindByEmail(req.Email)
	if existing != nil {
		return RegisterResponse{}, apperrors.NewConflict("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, apperrors.NewInternalServer("failed to hash password")
	}

	user, err := u.repo.Create(domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return RegisterResponse{}, err
	}

	token, err := jwtutil.Generate(user.ID, user.Role, user.Email)
	if err != nil {
		return RegisterResponse{}, err
	}

	return RegisterResponse{
		User: UserResponse{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Role:   user.Role,
			Status: user.Status,
		},
		Token: token,
	}, nil
}
