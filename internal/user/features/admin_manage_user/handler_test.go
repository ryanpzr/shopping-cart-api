package adminmanageuser_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adminmanageuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_manage_user"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) BanUser(id int) (adminmanageuser.ManageUserResponse, error) {
	args := m.Called(id)
	return args.Get(0).(adminmanageuser.ManageUserResponse), args.Error(1)
}
func (m *mockUsecase) TimeoutUser(id, hours int) (adminmanageuser.ManageUserResponse, error) {
	args := m.Called(id, hours)
	return args.Get(0).(adminmanageuser.ManageUserResponse), args.Error(1)
}
func (m *mockUsecase) UnbanUser(id int) (adminmanageuser.ManageUserResponse, error) {
	args := m.Called(id)
	return args.Get(0).(adminmanageuser.ManageUserResponse), args.Error(1)
}

func newRouter(h adminmanageuser.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/admin/users/:id/ban", h.BanUser)
	r.PATCH("/admin/users/:id/timeout", h.TimeoutUser)
	r.PATCH("/admin/users/:id/unban", h.UnbanUser)
	return r
}

func patchJSON(r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestBanHandler_InvalidID(t *testing.T) {
	uc := new(mockUsecase)
	r := newRouter(adminmanageuser.NewHandler(uc))
	w := patchJSON(r, "/admin/users/abc/ban", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBanHandler_TargetIsAdmin(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("BanUser", 1).Return(adminmanageuser.ManageUserResponse{}, apperrors.NewForbidden("cannot manage an admin user"))
	r := newRouter(adminmanageuser.NewHandler(uc))
	w := patchJSON(r, "/admin/users/1/ban", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestBanHandler_Success(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("BanUser", 2).Return(adminmanageuser.ManageUserResponse{ID: 2, Status: "banned"}, nil)
	r := newRouter(adminmanageuser.NewHandler(uc))
	w := patchJSON(r, "/admin/users/2/ban", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestTimeoutHandler_ZeroDuration(t *testing.T) {
	uc := new(mockUsecase)
	r := newRouter(adminmanageuser.NewHandler(uc))
	w := patchJSON(r, "/admin/users/2/timeout", map[string]int{"duration_hours": 0})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTimeoutHandler_NegativeDuration(t *testing.T) {
	uc := new(mockUsecase)
	r := newRouter(adminmanageuser.NewHandler(uc))
	w := patchJSON(r, "/admin/users/2/timeout", map[string]int{"duration_hours": -5})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTimeoutHandler_Success(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("TimeoutUser", 2, 3).Return(adminmanageuser.ManageUserResponse{ID: 2, Status: "timeout"}, nil)
	r := newRouter(adminmanageuser.NewHandler(uc))
	w := patchJSON(r, "/admin/users/2/timeout", map[string]int{"duration_hours": 3})
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestUnbanHandler_Success(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("UnbanUser", 2).Return(adminmanageuser.ManageUserResponse{ID: 2, Status: "active"}, nil)
	r := newRouter(adminmanageuser.NewHandler(uc))
	w := patchJSON(r, "/admin/users/2/unban", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}
