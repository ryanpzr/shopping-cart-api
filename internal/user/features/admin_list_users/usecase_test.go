package adminlistusers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	adminlistusers "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_list_users"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) FindAll(limit, offset int) ([]domain.User, int, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}
func (m *mockRepo) FindByID(id int) (*domain.User, error)                { return nil, nil }
func (m *mockRepo) FindByEmail(email string) (*domain.User, error)       { return nil, nil }
func (m *mockRepo) Create(u domain.User) (domain.User, error)            { return domain.User{}, nil }
func (m *mockRepo) UpdateProfile(id int, name, email string) (domain.User, error) {
	return domain.User{}, nil
}
func (m *mockRepo) UpdateStatus(id int, status string, t *time.Time) (domain.User, error) {
	return domain.User{}, nil
}

func TestListUsersUsecase_PaginationOffset(t *testing.T) {
	users := []domain.User{{ID: 3}, {ID: 4}}
	repo := new(mockRepo)
	// page=2, limit=2 → offset=2
	repo.On("FindAll", 2, 2).Return(users, 10, nil)

	uc := adminlistusers.NewUsecase(repo)
	resp, err := uc.ListUsers(2, 2)

	require.NoError(t, err)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 2, resp.Limit)
	assert.Equal(t, 10, resp.Total)
	assert.Equal(t, 5, resp.TotalPages) // ceil(10/2)
	assert.Len(t, resp.Data, 2)
	repo.AssertExpectations(t)
}

func TestListUsersUsecase_TotalPagesCeiling(t *testing.T) {
	repo := new(mockRepo)
	// total=11, limit=5 → total_pages=3 (ceil)
	repo.On("FindAll", 5, 0).Return(make([]domain.User, 5), 11, nil)

	uc := adminlistusers.NewUsecase(repo)
	resp, err := uc.ListUsers(1, 5)

	require.NoError(t, err)
	assert.Equal(t, 3, resp.TotalPages)
}

func TestListUsersUsecase_EmptyResult(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindAll", 20, 0).Return([]domain.User{}, 0, nil)

	uc := adminlistusers.NewUsecase(repo)
	resp, err := uc.ListUsers(1, 20)

	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Equal(t, 0, resp.TotalPages)
	assert.Empty(t, resp.Data)
}
