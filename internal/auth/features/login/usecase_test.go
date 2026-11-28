package login_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/outbox"
	outboxmocks "github.com/ryanpzr/shopping-cart-api/pkg/outbox/mocks"
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

func (m *mockUserRepo) FindByID(id int) (*domain.User, error) { return nil, nil }
func (m *mockUserRepo) FindAll(limit, offset int) ([]domain.User, int, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) UpdateProfile(id int, name, email string) (domain.User, error) {
	return domain.User{}, nil
}
func (m *mockUserRepo) UpdateStatus(id int, status string, t *time.Time) (domain.User, error) {
	return domain.User{}, nil
}

func hashedPassword(t *testing.T, pwd string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	require.NoError(t, err)
	return string(h)
}

func TestLoginUsecase_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	repo, events := new(mockUserRepo), new(outboxmocks.Writer)
	repo.On("FindByEmail", "user@example.com").Return(&domain.User{
		ID:           1,
		Name:         "User",
		Email:        "user@example.com",
		PasswordHash: hashedPassword(t, "password123"),
		Role:         "client",
		Status:       "active",
	}, nil)

	events.On("Insert", mock.Anything, outboxmocks.Matches(outbox.EventLogin, 1, nil)).Return(nil)
	uc := login.NewUsecase(repo, events)
	resp, err := uc.Login(context.Background(), login.LoginRequest{Email: "user@example.com", Password: "password123"})

	require.NoError(t, err)
	assert.Equal(t, "user@example.com", resp.User.Email)
	assert.NotEmpty(t, resp.Token)
	events.AssertExpectations(t)
}

func TestLoginUsecase_EmailNotFound(t *testing.T) {
	repo, events := new(mockUserRepo), new(outboxmocks.Writer)
	repo.On("FindByEmail", "noone@example.com").Return(nil, apperrors.NewNotFound("not found"))

	uc := login.NewUsecase(repo, events)
	_, err := uc.Login(context.Background(), login.LoginRequest{Email: "noone@example.com", Password: "password123"})

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Code)
	events.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func TestLoginUsecase_WrongPassword(t *testing.T) {
	repo, events := new(mockUserRepo), new(outboxmocks.Writer)
	repo.On("FindByEmail", "user@example.com").Return(&domain.User{
		PasswordHash: hashedPassword(t, "correct"),
		Status:       "active",
	}, nil)

	uc := login.NewUsecase(repo, events)
	_, err := uc.Login(context.Background(), login.LoginRequest{Email: "user@example.com", Password: "wrong"})

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Code)
	events.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func TestLoginUsecase_BannedUser(t *testing.T) {
	repo, events := new(mockUserRepo), new(outboxmocks.Writer)
	repo.On("FindByEmail", "banned@example.com").Return(&domain.User{
		PasswordHash: hashedPassword(t, "password123"),
		Status:       "banned",
	}, nil)

	uc := login.NewUsecase(repo, events)
	_, err := uc.Login(context.Background(), login.LoginRequest{Email: "banned@example.com", Password: "password123"})

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
	events.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func TestLoginUsecase_ActiveTimeout(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	repo, events := new(mockUserRepo), new(outboxmocks.Writer)
	repo.On("FindByEmail", "timeout@example.com").Return(&domain.User{
		PasswordHash: hashedPassword(t, "password123"),
		Status:       "timeout",
		TimeoutUntil: &future,
	}, nil)

	uc := login.NewUsecase(repo, events)
	_, err := uc.Login(context.Background(), login.LoginRequest{Email: "timeout@example.com", Password: "password123"})

	require.Error(t, err)
	var appErr *apperrors.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
	events.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func TestLoginUsecase_ExpiredTimeout(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	past := time.Now().Add(-1 * time.Hour)
	repo, events := new(mockUserRepo), new(outboxmocks.Writer)
	repo.On("FindByEmail", "expired@example.com").Return(&domain.User{
		ID:           2,
		Email:        "expired@example.com",
		PasswordHash: hashedPassword(t, "password123"),
		Role:         "client",
		Status:       "timeout",
		TimeoutUntil: &past,
	}, nil)

	events.On("Insert", mock.Anything, outboxmocks.Matches(outbox.EventLogin, 2, nil)).Return(nil)
	uc := login.NewUsecase(repo, events)
	resp, err := uc.Login(context.Background(), login.LoginRequest{Email: "expired@example.com", Password: "password123"})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
}

func TestLoginUsecase_FailingToWriteTheEventFailsTheLogin(t *testing.T) {
	// An action without its audit event is not allowed: no token is handed out.
	t.Setenv("JWT_SECRET", "test-secret")
	repo, events := new(mockUserRepo), new(outboxmocks.Writer)
	repo.On("FindByEmail", "user@example.com").Return(&domain.User{
		ID: 1, Email: "user@example.com", PasswordHash: hashedPassword(t, "password123"), Role: "client", Status: "active",
	}, nil)
	events.On("Insert", mock.Anything, mock.Anything).Return(errors.New("outbox: insert event: connection reset"))

	resp, err := login.NewUsecase(repo, events).Login(context.Background(),
		login.LoginRequest{Email: "user@example.com", Password: "password123"})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Code)
	assert.Empty(t, resp.Token)
}
