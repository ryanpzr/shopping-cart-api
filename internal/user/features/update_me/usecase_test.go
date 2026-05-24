package updateme_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	updateme "github.com/ryanpzr/shopping-cart-api/internal/user/features/update_me"
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
func (m *mockRepo) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockRepo) Create(u domain.User) (domain.User, error) { return domain.User{}, nil }
func (m *mockRepo) FindAll(limit, offset int) ([]domain.User, int, error) { return nil, 0, nil }
func (m *mockRepo) UpdateProfile(id int, name, email string) (domain.User, error) {
	args := m.Called(id, name, email)
	return args.Get(0).(domain.User), args.Error(1)
}
func (m *mockRepo) UpdateStatus(id int, status string, t *time.Time) (domain.User, error) {
	return domain.User{}, nil
}

var currentUser = &domain.User{
	ID:     1,
	Name:   "Alice",
	Email:  "alice@example.com",
	Role:   "client",
	Status: "active",
}

func TestUpdateMeUsecase_PartialUpdateNameOnly(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 1).Return(currentUser, nil)
	repo.On("UpdateProfile", 1, "NewName", "alice@example.com").Return(domain.User{
		ID: 1, Name: "NewName", Email: "alice@example.com", Role: "client", Status: "active",
	}, nil)

	uc := updateme.NewUsecase(repo)
	resp, err := uc.UpdateMe(1, updateme.UpdateMeRequest{Name: "NewName"})

	require.NoError(t, err)
	assert.Equal(t, "NewName", resp.Name)
	assert.Equal(t, "alice@example.com", resp.Email)
	repo.AssertExpectations(t)
}

func TestUpdateMeUsecase_EmailConflict(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 1).Return(currentUser, nil)
	repo.On("FindByEmail", "taken@example.com").Return(&domain.User{Email: "taken@example.com"}, nil)

	uc := updateme.NewUsecase(repo)
	_, err := uc.UpdateMe(1, updateme.UpdateMeRequest{Email: "taken@example.com"})

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)
}

func TestUpdateMeUsecase_SameEmailNoConflictCheck(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 1).Return(currentUser, nil)
	// FindByEmail should NOT be called when email hasn't changed
	repo.On("UpdateProfile", 1, "Alice", "alice@example.com").Return(domain.User{
		ID: 1, Name: "Alice", Email: "alice@example.com", Role: "client", Status: "active",
	}, nil)

	uc := updateme.NewUsecase(repo)
	_, err := uc.UpdateMe(1, updateme.UpdateMeRequest{Email: "alice@example.com"})

	require.NoError(t, err)
	repo.AssertNotCalled(t, "FindByEmail", mock.Anything)
}
