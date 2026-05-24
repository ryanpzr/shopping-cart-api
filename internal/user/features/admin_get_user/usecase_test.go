package admingetuser_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	admingetuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_get_user"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) FindByID(id int) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockRepo) FindByEmail(email string) (*domain.User, error)        { return nil, nil }
func (m *mockRepo) Create(u domain.User) (domain.User, error)             { return domain.User{}, nil }
func (m *mockRepo) FindAll(limit, offset int) ([]domain.User, int, error) { return nil, 0, nil }
func (m *mockRepo) UpdateProfile(id int, name, email string) (domain.User, error) {
	return domain.User{}, nil
}
func (m *mockRepo) UpdateStatus(id int, status string, t *time.Time) (domain.User, error) {
	return domain.User{}, nil
}

func TestGetUserUsecase_Success(t *testing.T) {
	now := time.Now()
	repo := new(mockRepo)
	repo.On("FindByID", 5).Return(&domain.User{
		ID: 5, Name: "Bob", Email: "bob@example.com",
		Role: "client", Status: "active", CreatedAt: now, UpdatedAt: now,
	}, nil)

	uc := admingetuser.NewUsecase(repo)
	resp, err := uc.GetUser(5)

	require.NoError(t, err)
	assert.Equal(t, 5, resp.ID)
	assert.Equal(t, "Bob", resp.Name)
	repo.AssertExpectations(t)
}

func TestGetUserUsecase_NotFound(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 999).Return(nil, apperrors.NewNotFound("user not found"))

	uc := admingetuser.NewUsecase(repo)
	_, err := uc.GetUser(999)

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Code)
}
