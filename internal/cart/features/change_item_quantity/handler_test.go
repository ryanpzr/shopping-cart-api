package changeitemquantity_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	changeitemquantity "github.com/ryanpzr/shopping-cart-api/internal/cart/features/change_item_quantity"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) ChangeQuantity(ctx context.Context, buyerID, itemID int, req changeitemquantity.ChangeQuantityRequest) (changeitemquantity.CartResponse, error) {
	args := m.Called(ctx, buyerID, itemID, req)
	return args.Get(0).(changeitemquantity.CartResponse), args.Error(1)
}

func patch(uc changeitemquantity.Usecase, path, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/carts/me/items/:itemId", func(c *gin.Context) {
		c.Set("user_id", buyer)
		changeitemquantity.NewHandler(uc).ChangeQuantity(c)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestChangeQuantityHandler_OK(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("ChangeQuantity", mock.Anything, buyer, 10, mock.MatchedBy(func(r changeitemquantity.ChangeQuantityRequest) bool {
		return *r.Quantity == 3
	})).Return(changeitemquantity.CartResponse{ID: 1}, nil)

	w := patch(uc, "/carts/me/items/10", `{"quantity":3}`)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestChangeQuantityHandler_BadInput(t *testing.T) {
	cases := map[string]string{
		"/carts/me/items/abc": `{"quantity":1}`,
		"/carts/me/items/0":   `{"quantity":1}`,
		"/carts/me/items/10":  `{`,
	}
	for path, body := range cases {
		uc := new(mockUsecase)
		w := patch(uc, path, body)
		assert.Equal(t, http.StatusBadRequest, w.Code, path)
		uc.AssertNotCalled(t, "ChangeQuantity", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	}
}
