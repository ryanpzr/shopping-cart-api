package adminactivitylog_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/activitylog/domain"
	adminactivitylog "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_activity_log"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockRepo struct{ mock.Mock }

func (m *mockRepo) FindByUser(ctx context.Context, userID int, eventType string, limit, offset int) ([]domain.Entry, int64, error) {
	args := m.Called(ctx, userID, eventType, limit, offset)
	entries, _ := args.Get(0).([]domain.Entry)
	return entries, int64(args.Int(1)), args.Error(2)
}

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func entry(id, eventType string) domain.Entry {
	return domain.Entry{
		EventID:    id,
		UserID:     42,
		EventType:  eventType,
		Metadata:   map[string]any{"order_id": 10},
		OccurredAt: time.Date(2026, 11, 28, 10, 30, 0, 0, time.UTC),
	}
}

func TestGetActivityLog_MapsEntriesAndPages(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByUser", mock.Anything, 42, "", 20, 0).
		Return([]domain.Entry{entry("e1", "cart_checkout"), entry("e2", "order_placed")}, 45, nil)

	resp, err := adminactivitylog.NewUsecase(repo).GetActivityLog(context.Background(), 42, 1, 20, "")

	require.NoError(t, err)
	require.Len(t, resp.Data, 2)
	assert.Equal(t, "e1", resp.Data[0].ID)
	assert.Equal(t, 42, resp.Data[0].UserID)
	assert.Equal(t, "cart_checkout", resp.Data[0].EventType)
	assert.EqualValues(t, 10, resp.Data[0].Metadata["order_id"])
	assert.Equal(t, 45, resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.Limit)
	assert.Equal(t, 3, resp.TotalPages, "45 entries in pages of 20")
}

func TestGetActivityLog_OffsetFollowsThePage(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByUser", mock.Anything, 42, "", 10, 30).Return([]domain.Entry{}, 33, nil)

	resp, err := adminactivitylog.NewUsecase(repo).GetActivityLog(context.Background(), 42, 4, 10, "")

	require.NoError(t, err)
	assert.Equal(t, 4, resp.Page)
	assert.Equal(t, 4, resp.TotalPages)
	repo.AssertExpectations(t)
}

func TestGetActivityLog_EmptyIsAnEmptyArray(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByUser", mock.Anything, 42, "", 20, 0).Return(nil, 0, nil)

	resp, err := adminactivitylog.NewUsecase(repo).GetActivityLog(context.Background(), 42, 1, 20, "")

	require.NoError(t, err)
	assert.NotNil(t, resp.Data)
	assert.Empty(t, resp.Data)
	assert.Equal(t, 0, resp.TotalPages)
}

func TestGetActivityLog_NilMetadataBecomesAnEmptyObject(t *testing.T) {
	repo := new(mockRepo)
	e := entry("e1", "login")
	e.Metadata = nil
	repo.On("FindByUser", mock.Anything, 42, "", 20, 0).Return([]domain.Entry{e}, 1, nil)

	resp, err := adminactivitylog.NewUsecase(repo).GetActivityLog(context.Background(), 42, 1, 20, "")

	require.NoError(t, err)
	assert.NotNil(t, resp.Data[0].Metadata)
}

func TestGetActivityLog_FiltersByEventType(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByUser", mock.Anything, 42, "order_cancelled", 20, 0).
		Return([]domain.Entry{entry("e1", "order_cancelled")}, 1, nil)

	resp, err := adminactivitylog.NewUsecase(repo).GetActivityLog(context.Background(), 42, 1, 20, "order_cancelled")

	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	repo.AssertExpectations(t)
}

func TestGetActivityLog_UnknownEventTypeIsABadRequest(t *testing.T) {
	for _, eventType := range []string{"user_banned", "LOGIN", " login"} {
		repo := new(mockRepo)

		_, err := adminactivitylog.NewUsecase(repo).GetActivityLog(context.Background(), 42, 1, 20, eventType)

		assertAppError(t, err, http.StatusBadRequest)
		assert.ErrorContains(t, err, "login", "the message lists the valid types")
		repo.AssertNotCalled(t, "FindByUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	}
}

func TestGetActivityLog_StoreDownIs503WithAFixedMessage(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByUser", mock.Anything, 42, "", 20, 0).
		Return(nil, 0, fmt.Errorf("%w: server selection timeout, 10.0.0.5:27017", domain.ErrStoreUnavailable))

	_, err := adminactivitylog.NewUsecase(repo).GetActivityLog(context.Background(), 42, 1, 20, "")

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusServiceUnavailable, appErr.Code)
	assert.Equal(t, "activity log is temporarily unavailable", appErr.Message, "no driver detail reaches the client")
}

func TestGetActivityLog_OtherFailuresAre500(t *testing.T) {
	repo := new(mockRepo)
	repo.On("FindByUser", mock.Anything, 42, "", 20, 0).Return(nil, 0, errors.New("decode failure"))

	_, err := adminactivitylog.NewUsecase(repo).GetActivityLog(context.Background(), 42, 1, 20, "")

	assertAppError(t, err, http.StatusInternalServerError)
}
