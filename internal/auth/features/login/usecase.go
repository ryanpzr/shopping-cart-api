package login

import (
	"time"

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

func (u *usecase) Login(req LoginRequest) (LoginResponse, error) {
	user, err := u.repo.FindByEmail(req.Email)
	if err != nil {
		return LoginResponse{}, apperrors.NewUnauthorized("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return LoginResponse{}, apperrors.NewUnauthorized("invalid credentials")
	}

	if user.Status == "banned" {
		return LoginResponse{}, apperrors.NewForbidden("user is banned")
	}
	if user.Status == "timeout" && user.TimeoutUntil != nil && user.TimeoutUntil.After(time.Now()) {
		return LoginResponse{}, apperrors.NewForbidden("user is in timeout")
	}

	token, err := jwtutil.Generate(user.ID, user.Role, user.Email)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
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
