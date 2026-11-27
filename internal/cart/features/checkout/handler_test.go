package checkout_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/features/checkout"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Checkout(ctx context.Context, buyerID int) (checkout.OrderResponse, error) {
	args := m.Called(ctx, buyerID)
	return args.Get(0).(checkout.OrderResponse), args.Error(1)
}

func post(uc checkout.Usecase) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/carts/me/checkout", func(c *gin.Context) { c.Set("user_id", buyer); checkout.NewHandler(uc).Checkout(c) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/carts/me/checkout", nil))
	return w
}

func TestCheckoutHandler_Created(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Checkout", mock.Anything, buyer).Return(checkout.OrderResponse{ID: 55, Status: "pending", TotalPrice: money.FromCents(38484)}, nil)

	w := post(uc)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"total_price":384.84`)
}

func TestCheckoutHandler_Conflict(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Checkout", mock.Anything, buyer).Return(checkout.OrderResponse{}, apperrors.NewConflict("cart is empty"))

	w := post(uc)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.JSONEq(t, `{"error":"cart is empty"}`, w.Body.String())
}
