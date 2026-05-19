package register_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) Create(user domain.User) (domain.User, error) {
	args := m.Called(user)
	return args.Get(0).(domain.User), args.Error(1)
}

func TestRegisterUsecase_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	repo := new(mockUserRepo)
	repo.On("FindByEmail", "new@example.com").Return(nil, apperrors.NewNotFound("not found"))
	repo.On("Create", mock.Anything).Return(domain.User{
		ID:     1,
		Name:   "New User",
		Email:  "new@example.com",
		Role:   "client",
		Status: "active",
	}, nil)

	uc := register.NewUsecase(repo)
	resp, err := uc.Register(register.RegisterRequest{
		Name:     "New User",
		Email:    "new@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, "new@example.com", resp.User.Email)
	assert.Equal(t, "client", resp.User.Role)
	assert.NotEmpty(t, resp.Token)
	repo.AssertExpectations(t)
}

func TestRegisterUsecase_EmailConflict(t *testing.T) {
	repo := new(mockUserRepo)
	repo.On("FindByEmail", "taken@example.com").Return(&domain.User{Email: "taken@example.com"}, nil)

	uc := register.NewUsecase(repo)
	_, err := uc.Register(register.RegisterRequest{
		Name:     "User",
		Email:    "taken@example.com",
		Password: "password123",
	})

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)
}
