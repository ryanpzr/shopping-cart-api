package getcart_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	getcart "github.com/ryanpzr/shopping-cart-api/internal/cart/features/get_cart"
	"github.com/ryanpzr/shopping-cart-api/pkg/money"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Get(ctx context.Context, buyerID int) (getcart.CartResponse, error) {
	args := m.Called(ctx, buyerID)
	return args.Get(0).(getcart.CartResponse), args.Error(1)
}

func TestGetHandler_UsesAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := new(mockUsecase)
	uc.On("Get", mock.Anything, 7).Return(getcart.CartResponse{
		ID: 1, Status: "active", Items: []getcart.ItemResponse{}, Subtotal: money.FromCents(17980), Total: money.FromCents(17980),
	}, nil)
	r := gin.New()
	r.GET("/carts/me", func(c *gin.Context) { c.Set("user_id", 7); getcart.NewHandler(uc).Get(c) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/carts/me", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"id":1,"status":"active","items":[],"subtotal":179.80,"discount_total":0.00,"total":179.80}`, w.Body.String())
}
