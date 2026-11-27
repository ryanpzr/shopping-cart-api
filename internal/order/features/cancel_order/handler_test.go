package cancelorder_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	cancelorder "github.com/ryanpzr/shopping-cart-api/internal/order/features/cancel_order"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Cancel(ctx context.Context, orderID, userID int) (cancelorder.OrderResponse, error) {
	args := m.Called(ctx, orderID, userID)
	return args.Get(0).(cancelorder.OrderResponse), args.Error(1)
}

func patch(uc cancelorder.Usecase, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/orders/:id/cancel", func(c *gin.Context) { c.Set("user_id", owner); cancelorder.NewHandler(uc).Cancel(c) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, path, nil))
	return w
}

func TestCancelHandler_OK(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Cancel", mock.Anything, orderID, owner).Return(cancelorder.OrderResponse{ID: orderID, Status: "cancelled"}, nil)

	w := patch(uc, "/orders/55/cancel")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"cancelled"`)
}

func TestCancelHandler_InvalidID(t *testing.T) {
	uc := new(mockUsecase)
	assert.Equal(t, http.StatusBadRequest, patch(uc, "/orders/abc/cancel").Code)
	assert.Equal(t, http.StatusBadRequest, patch(uc, "/orders/0/cancel").Code)
	uc.AssertNotCalled(t, "Cancel", mock.Anything, mock.Anything, mock.Anything)
}

func TestCancelHandler_ErrorCodes(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Cancel", mock.Anything, 1, owner).Return(cancelorder.OrderResponse{}, apperrors.NewForbidden("not yours"))
	uc.On("Cancel", mock.Anything, 2, owner).Return(cancelorder.OrderResponse{}, apperrors.NewConflict("only a pending order can be cancelled"))
	uc.On("Cancel", mock.Anything, 3, owner).Return(cancelorder.OrderResponse{}, apperrors.NewNotFound("order not found"))

	assert.Equal(t, http.StatusForbidden, patch(uc, "/orders/1/cancel").Code)
	assert.Equal(t, http.StatusConflict, patch(uc, "/orders/2/cancel").Code)
	assert.Equal(t, http.StatusNotFound, patch(uc, "/orders/3/cancel").Code)
}
