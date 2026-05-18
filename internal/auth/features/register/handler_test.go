package register_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockRegisterUsecase struct {
	mock.Mock
}

func (m *mockRegisterUsecase) Register(req register.RegisterRequest) (register.RegisterResponse, error) {
	args := m.Called(req)
	return args.Get(0).(register.RegisterResponse), args.Error(1)
}

func newRegisterRouter(h register.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/register", h.Register)
	return r
}

func postJSON(r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	uc := new(mockRegisterUsecase)
	r := newRegisterRouter(register.NewHandler(uc))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_MissingName(t *testing.T) {
	uc := new(mockRegisterUsecase)
	r := newRegisterRouter(register.NewHandler(uc))

	w := postJSON(r, "/auth/register", map[string]string{"email": "a@b.com", "password": "password123"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_InvalidEmail(t *testing.T) {
	uc := new(mockRegisterUsecase)
	r := newRegisterRouter(register.NewHandler(uc))

	w := postJSON(r, "/auth/register", map[string]string{"name": "Test", "email": "notanemail", "password": "password123"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_ShortPassword(t *testing.T) {
	uc := new(mockRegisterUsecase)
	r := newRegisterRouter(register.NewHandler(uc))

	w := postJSON(r, "/auth/register", map[string]string{"name": "Test", "email": "a@b.com", "password": "short"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_Success(t *testing.T) {
	uc := new(mockRegisterUsecase)
	uc.On("Register", mock.Anything).Return(register.RegisterResponse{
		User:  register.UserResponse{ID: 1, Email: "a@b.com"},
		Token: "jwt-token",
	}, nil)
	r := newRegisterRouter(register.NewHandler(uc))

	w := postJSON(r, "/auth/register", map[string]string{"name": "Test", "email": "a@b.com", "password": "password123"})
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRegisterHandler_Conflict(t *testing.T) {
	uc := new(mockRegisterUsecase)
	uc.On("Register", mock.Anything).Return(register.RegisterResponse{}, apperrors.NewConflict("email already registered"))
	r := newRegisterRouter(register.NewHandler(uc))

	w := postJSON(r, "/auth/register", map[string]string{"name": "Test", "email": "a@b.com", "password": "password123"})
	assert.Equal(t, http.StatusConflict, w.Code)
}
