package adminlistusers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	adminlistusers "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_list_users"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) ListUsers(page, limit int) (adminlistusers.PaginatedResponse, error) {
	args := m.Called(page, limit)
	return args.Get(0).(adminlistusers.PaginatedResponse), args.Error(1)
}

func newRouter(h adminlistusers.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/users", h.ListUsers)
	return r
}

func TestListUsersHandler_Defaults(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("ListUsers", 1, 20).Return(adminlistusers.PaginatedResponse{
		Data: []adminlistusers.UserSummary{}, Total: 0, Page: 1, Limit: 20, TotalPages: 0,
	}, nil)

	r := newRouter(adminlistusers.NewHandler(uc))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestListUsersHandler_CustomPagination(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("ListUsers", 2, 5).Return(adminlistusers.PaginatedResponse{
		Data: []adminlistusers.UserSummary{}, Total: 50, Page: 2, Limit: 5, TotalPages: 10,
	}, nil)

	r := newRouter(adminlistusers.NewHandler(uc))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users?page=2&limit=5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body adminlistusers.PaginatedResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 2, body.Page)
	assert.Equal(t, 5, body.Limit)
	uc.AssertExpectations(t)
}

func TestListUsersHandler_LimitCappedAt100(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("ListUsers", 1, 100).Return(adminlistusers.PaginatedResponse{
		Data: []adminlistusers.UserSummary{},
	}, nil)

	r := newRouter(adminlistusers.NewHandler(uc))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users?limit=999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}
