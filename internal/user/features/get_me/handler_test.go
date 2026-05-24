package getme_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	getme "github.com/ryanpzr/shopping-cart-api/internal/user/features/get_me"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) GetMe(userID int) (getme.GetMeResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(getme.GetMeResponse), args.Error(1)
}

func newRouter(h getme.Handler, userID interface{}) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/me", func(c *gin.Context) {
		if userID != nil {
			c.Set("user_id", userID)
		}
		h.GetMe(c)
	})
	return r
}

func TestGetMeHandler_Success(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("GetMe", 1).Return(getme.GetMeResponse{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Role:      "client",
		Status:    "active",
		CreatedAt: time.Now(),
	}, nil)

	r := newRouter(getme.NewHandler(uc), 1)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(1), body["id"])
	assert.Equal(t, "alice@example.com", body["email"])
	assert.Nil(t, body["password_hash"])
	uc.AssertExpectations(t)
}

func TestGetMeHandler_MissingUserID(t *testing.T) {
	uc := new(mockUsecase)
	r := newRouter(getme.NewHandler(uc), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMeHandler_NotFound(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("GetMe", 99).Return(getme.GetMeResponse{}, apperrors.NewNotFound("user not found"))

	r := newRouter(getme.NewHandler(uc), 99)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
