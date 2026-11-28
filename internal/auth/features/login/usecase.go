package login

import (
	"context"
	"time"

	usershared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	jwtutil "github.com/ryanpzr/shopping-cart-api/pkg/jwt"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
	"golang.org/x/crypto/bcrypt"
)

type usecase struct {
	events outbox.Writer
	repo   usershared.Repository
}

func NewUsecase(repo usershared.Repository, events outbox.Writer) Usecase {
	return &usecase{repo: repo, events: events}
}

// Login checks the credentials and the account status, then records a login
// audit event before handing out the token. If the event cannot be written the
// login fails: an action without its audit event is not allowed. The event is
// a single INSERT, so it is atomic on its own.
func (u *usecase) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
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

	if err := u.events.Insert(ctx, outbox.New(user.ID, outbox.EventLogin, nil)); err != nil {
		return LoginResponse{}, apperrors.NewInternalServer(err.Error())
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
