package removeitem_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	removeitem "github.com/ryanpzr/shopping-cart-api/internal/cart/features/remove_item"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) RemoveItem(ctx context.Context, buyerID, itemID int) (removeitem.CartResponse, error) {
	args := m.Called(ctx, buyerID, itemID)
	return args.Get(0).(removeitem.CartResponse), args.Error(1)
}

func del(uc removeitem.Usecase, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/carts/me/items/:itemId", func(c *gin.Context) { c.Set("user_id", 7); removeitem.NewHandler(uc).RemoveItem(c) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, path, nil))
	return w
}

func TestRemoveItemHandler(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("RemoveItem", mock.Anything, 7, 10).Return(removeitem.CartResponse{ID: 1}, nil)
	uc.On("RemoveItem", mock.Anything, 7, 99).Return(removeitem.CartResponse{}, apperrors.NewNotFound("cart item not found"))

	assert.Equal(t, http.StatusOK, del(uc, "/carts/me/items/10").Code)
	assert.Equal(t, http.StatusNotFound, del(uc, "/carts/me/items/99").Code)
	assert.Equal(t, http.StatusBadRequest, del(uc, "/carts/me/items/x").Code)
}
