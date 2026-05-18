package login_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockLoginUsecase struct {
	mock.Mock
}

func (m *mockLoginUsecase) Login(req login.LoginRequest) (login.LoginResponse, error) {
	args := m.Called(req)
	return args.Get(0).(login.LoginResponse), args.Error(1)
}

func newLoginRouter(h login.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", h.Login)
	return r
}

func postLoginJSON(r *gin.Engine, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestLoginHandler_InvalidBody(t *testing.T) {
	uc := new(mockLoginUsecase)
	r := newLoginRouter(login.NewHandler(uc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_MissingFields(t *testing.T) {
	uc := new(mockLoginUsecase)
	r := newLoginRouter(login.NewHandler(uc))

	w := postLoginJSON(r, map[string]string{"email": "a@b.com"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	uc := new(mockLoginUsecase)
	uc.On("Login", mock.Anything).Return(login.LoginResponse{
		User:  login.UserResponse{ID: 1, Email: "a@b.com"},
		Token: "jwt-token",
	}, nil)
	r := newLoginRouter(login.NewHandler(uc))

	w := postLoginJSON(r, map[string]string{"email": "a@b.com", "password": "password123"})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginHandler_Unauthorized(t *testing.T) {
	uc := new(mockLoginUsecase)
	uc.On("Login", mock.Anything).Return(login.LoginResponse{}, apperrors.NewUnauthorized("invalid credentials"))
	r := newLoginRouter(login.NewHandler(uc))

	w := postLoginJSON(r, map[string]string{"email": "a@b.com", "password": "wrong"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginHandler_Forbidden(t *testing.T) {
	uc := new(mockLoginUsecase)
	uc.On("Login", mock.Anything).Return(login.LoginResponse{}, apperrors.NewForbidden("user is banned"))
	r := newLoginRouter(login.NewHandler(uc))

	w := postLoginJSON(r, map[string]string{"email": "a@b.com", "password": "password123"})
	assert.Equal(t, http.StatusForbidden, w.Code)
}
