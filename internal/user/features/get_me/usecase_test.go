package getme_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	getme "github.com/ryanpzr/shopping-cart-api/internal/user/features/get_me"
	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
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
func (m *mockRepo) FindByEmail(email string) (*domain.User, error)       { return nil, nil }
func (m *mockRepo) Create(u domain.User) (domain.User, error)            { return domain.User{}, nil }
func (m *mockRepo) FindAll(limit, offset int) ([]domain.User, int, error) { return nil, 0, nil }
func (m *mockRepo) UpdateProfile(id int, name, email string) (domain.User, error) {
	return domain.User{}, nil
}
func (m *mockRepo) UpdateStatus(id int, status string, t *time.Time) (domain.User, error) {
	return domain.User{}, nil
}

func TestGetMeUsecase_Success(t *testing.T) {
	now := time.Now()
	repo := new(mockRepo)
	repo.On("FindByID", 1).Return(&domain.User{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Role:      "client",
		Status:    "active",
		CreatedAt: now,
	}, nil)

	uc := getme.NewUsecase(repo)
	resp, err := uc.GetMe(1)

	require.NoError(t, err)
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, "Alice", resp.Name)
	assert.Equal(t, "alice@example.com", resp.Email)
	assert.Empty(t, "") // password_hash not in response (structural guarantee)
	repo.AssertExpectations(t)
}

func TestGetMeUsecase_NotFound(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 99).Return(nil, apperrors.NewNotFound("user not found"))

	uc := getme.NewUsecase(repo)
	_, err := uc.GetMe(99)

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Code)
}
