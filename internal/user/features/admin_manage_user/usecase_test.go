package adminmanageuser_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	adminmanageuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_manage_user"
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
func (m *mockRepo) UpdateStatus(id int, status string, t *time.Time) (domain.User, error) {
	args := m.Called(id, status, t)
	return args.Get(0).(domain.User), args.Error(1)
}
func (m *mockRepo) FindByEmail(email string) (*domain.User, error)        { return nil, nil }
func (m *mockRepo) Create(u domain.User) (domain.User, error)             { return domain.User{}, nil }
func (m *mockRepo) FindAll(limit, offset int) ([]domain.User, int, error) { return nil, 0, nil }
func (m *mockRepo) UpdateProfile(id int, name, email string) (domain.User, error) {
	return domain.User{}, nil
}

var clientUser = &domain.User{ID: 2, Name: "Bob", Email: "bob@example.com", Role: "client", Status: "active"}
var adminUser = &domain.User{ID: 1, Name: "Admin", Email: "admin@example.com", Role: "admin", Status: "active"}

func TestBanUser_Success(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 2).Return(clientUser, nil)
	repo.On("UpdateStatus", 2, "banned", (*time.Time)(nil)).Return(
		domain.User{ID: 2, Status: "banned"}, nil,
	)

	uc := adminmanageuser.NewUsecase(repo)
	resp, err := uc.BanUser(2)

	require.NoError(t, err)
	assert.Equal(t, "banned", resp.Status)
	repo.AssertExpectations(t)
}

func TestBanUser_TargetIsAdmin(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 1).Return(adminUser, nil)

	uc := adminmanageuser.NewUsecase(repo)
	_, err := uc.BanUser(1)

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}

func TestTimeoutUser_Success(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 2).Return(clientUser, nil)
	repo.On("UpdateStatus", 2, "timeout", mock.AnythingOfType("*time.Time")).Return(
		domain.User{ID: 2, Status: "timeout"}, nil,
	)

	uc := adminmanageuser.NewUsecase(repo)
	resp, err := uc.TimeoutUser(2, 3)

	require.NoError(t, err)
	assert.Equal(t, "timeout", resp.Status)
	repo.AssertExpectations(t)
}

func TestTimeoutUser_TargetIsAdmin(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 1).Return(adminUser, nil)

	uc := adminmanageuser.NewUsecase(repo)
	_, err := uc.TimeoutUser(1, 2)

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}

func TestUnbanUser_Success(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByID", 2).Return(clientUser, nil)
	repo.On("UpdateStatus", 2, "active", (*time.Time)(nil)).Return(
		domain.User{ID: 2, Status: "active"}, nil,
	)

	uc := adminmanageuser.NewUsecase(repo)
	resp, err := uc.UnbanUser(2)

	require.NoError(t, err)
	assert.Equal(t, "active", resp.Status)
	repo.AssertExpectations(t)
}
