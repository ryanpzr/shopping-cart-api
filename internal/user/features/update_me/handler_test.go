package updateme_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	updateme "github.com/ryanpzr/shopping-cart-api/internal/user/features/update_me"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) UpdateMe(userID int, req updateme.UpdateMeRequest) (updateme.UpdateMeResponse, error) {
	args := m.Called(userID, req)
	return args.Get(0).(updateme.UpdateMeResponse), args.Error(1)
}

func newRouter(h updateme.Handler, userID interface{}) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/users/me", func(c *gin.Context) {
		if userID != nil {
			c.Set("user_id", userID)
		}
		h.UpdateMe(c)
	})
	return r
}

func putJSON(r *gin.Engine, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestUpdateMeHandler_InvalidEmail(t *testing.T) {
	uc := new(mockUsecase)
	r := newRouter(updateme.NewHandler(uc), 1)
	w := putJSON(r, map[string]string{"email": "notanemail"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateMeHandler_EmptyBody(t *testing.T) {
	uc := new(mockUsecase)
	r := newRouter(updateme.NewHandler(uc), 1)
	w := putJSON(r, map[string]string{})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateMeHandler_Success(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("UpdateMe", 1, updateme.UpdateMeRequest{Name: "NewName"}).Return(updateme.UpdateMeResponse{
		ID: 1, Name: "NewName", Email: "alice@example.com",
	}, nil)

	r := newRouter(updateme.NewHandler(uc), 1)
	w := putJSON(r, map[string]string{"name": "NewName"})
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestUpdateMeHandler_EmailConflict(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("UpdateMe", 1, updateme.UpdateMeRequest{Email: "taken@example.com"}).
		Return(updateme.UpdateMeResponse{}, apperrors.NewConflict("email already in use"))

	r := newRouter(updateme.NewHandler(uc), 1)
	w := putJSON(r, map[string]string{"email": "taken@example.com"})
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUpdateMeHandler_MissingUserID(t *testing.T) {
	uc := new(mockUsecase)
	r := newRouter(updateme.NewHandler(uc), nil)
	w := putJSON(r, map[string]string{"name": "Test"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
