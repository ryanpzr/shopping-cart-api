package admingetuser_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	admingetuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_get_user"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) GetUser(id int) (admingetuser.AdminUserResponse, error) {
	args := m.Called(id)
	return args.Get(0).(admingetuser.AdminUserResponse), args.Error(1)
}

func newRouter(h admingetuser.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/users/:id", h.GetUser)
	return r
}

func TestGetUserHandler_InvalidID(t *testing.T) {
	uc := new(mockUsecase)
	r := newRouter(admingetuser.NewHandler(uc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserHandler_NotFound(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("GetUser", 999).Return(admingetuser.AdminUserResponse{}, apperrors.NewNotFound("user not found"))
	r := newRouter(admingetuser.NewHandler(uc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserHandler_Success(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("GetUser", 5).Return(admingetuser.AdminUserResponse{
		ID: 5, Name: "Bob", Email: "bob@example.com", Role: "client", Status: "active",
	}, nil)
	r := newRouter(admingetuser.NewHandler(uc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users/5", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}
