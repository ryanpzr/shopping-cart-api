package adminactivitylog_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	adminactivitylog "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_activity_log"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) GetActivityLog(ctx context.Context, userID, page, limit int, eventType string) (adminactivitylog.ActivityLogResponse, error) {
	args := m.Called(ctx, userID, page, limit, eventType)
	return args.Get(0).(adminactivitylog.ActivityLogResponse), args.Error(1)
}

func get(uc adminactivitylog.Usecase, url string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/users/:id/activity", adminactivitylog.NewHandler(uc).GetActivityLog)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
	return w
}

func TestActivityLogHandler_ForwardsPathAndQuery(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("GetActivityLog", mock.Anything, 7, 3, 10, "login").
		Return(adminactivitylog.ActivityLogResponse{Data: []adminactivitylog.ActivityLogEntry{}, Page: 3, Limit: 10}, nil)

	w := get(uc, "/admin/users/7/activity?page=3&limit=10&event_type=login")

	assert.Equal(t, http.StatusOK, w.Code)
	var body adminactivitylog.ActivityLogResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 3, body.Page)
	assert.Equal(t, 10, body.Limit)
	uc.AssertExpectations(t)
}

func TestActivityLogHandler_Defaults(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("GetActivityLog", mock.Anything, 7, 1, 20, "").Return(adminactivitylog.ActivityLogResponse{}, nil)

	assert.Equal(t, http.StatusOK, get(uc, "/admin/users/7/activity").Code)
	uc.AssertExpectations(t)
}

func TestActivityLogHandler_InvalidPaginationFallsBackAndLimitIsCapped(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("GetActivityLog", mock.Anything, 7, 1, 20, "").Return(adminactivitylog.ActivityLogResponse{}, nil).Once()
	uc.On("GetActivityLog", mock.Anything, 7, 1, 100, "").Return(adminactivitylog.ActivityLogResponse{}, nil).Once()

	assert.Equal(t, http.StatusOK, get(uc, "/admin/users/7/activity?page=zero&limit=-4").Code)
	assert.Equal(t, http.StatusOK, get(uc, "/admin/users/7/activity?limit=9999").Code)
	uc.AssertExpectations(t)
}

func TestActivityLogHandler_InvalidUserID(t *testing.T) {
	for _, path := range []string{"/admin/users/abc/activity", "/admin/users/0/activity", "/admin/users/-2/activity"} {
		uc := new(mockUsecase)
		assert.Equal(t, http.StatusBadRequest, get(uc, path).Code, path)
		uc.AssertNotCalled(t, "GetActivityLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	}
}

func TestActivityLogHandler_UsecaseErrorsKeepTheirStatus(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("GetActivityLog", mock.Anything, 1, mock.Anything, mock.Anything, "bad").
		Return(adminactivitylog.ActivityLogResponse{}, apperrors.NewBadRequest("event_type must be one of: login"))
	uc.On("GetActivityLog", mock.Anything, 2, mock.Anything, mock.Anything, "").
		Return(adminactivitylog.ActivityLogResponse{}, apperrors.NewServiceUnavailable("activity log is temporarily unavailable"))

	assert.Equal(t, http.StatusBadRequest, get(uc, "/admin/users/1/activity?event_type=bad").Code)
	w := get(uc, "/admin/users/2/activity")
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"error":"activity log is temporarily unavailable"}`, w.Body.String())
}

func TestActivityLogHandler_HugePageIs400(t *testing.T) {
	uc := new(mockUsecase)
	w := get(uc, "/admin/users/1/activity?page=999999999999999999&limit=100")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "GetActivityLog")
}
