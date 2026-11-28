package getorder_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	getorder "github.com/ryanpzr/shopping-cart-api/internal/order/features/get_order"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Get(ctx context.Context, orderID, userID int, role string) (getorder.OrderResponse, error) {
	args := m.Called(ctx, orderID, userID, role)
	return args.Get(0).(getorder.OrderResponse), args.Error(1)
}

func get(uc getorder.Usecase, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/orders/:id", func(c *gin.Context) {
		c.Set("user_id", 7)
		c.Set("role", "client")
		getorder.NewHandler(uc).Get(c)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestGetHandler_PassesUserAndRole(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Get", mock.Anything, 55, 7, "client").Return(getorder.OrderResponse{ID: 55, Status: "paid"}, nil)

	w := get(uc, "/orders/55")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"paid"`)
}

func TestGetHandler_InvalidID(t *testing.T) {
	for _, path := range []string{"/orders/abc", "/orders/0", "/orders/-3"} {
		uc := new(mockUsecase)
		assert.Equal(t, http.StatusBadRequest, get(uc, path).Code, path)
		uc.AssertNotCalled(t, "Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	}
}

func TestGetHandler_ForbiddenAndNotFound(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Get", mock.Anything, 55, 7, "client").Return(getorder.OrderResponse{}, apperrors.NewForbidden("not yours"))
	uc.On("Get", mock.Anything, 99, 7, "client").Return(getorder.OrderResponse{}, apperrors.NewNotFound("order not found"))

	assert.Equal(t, http.StatusForbidden, get(uc, "/orders/55").Code)
	assert.Equal(t, http.StatusNotFound, get(uc, "/orders/99").Code)
}
