package additemtocart_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	additemtocart "github.com/ryanpzr/shopping-cart-api/internal/cart/features/add_item_to_cart"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) AddItem(ctx context.Context, buyerID int, req additemtocart.AddItemRequest) (additemtocart.CartResponse, error) {
	args := m.Called(ctx, buyerID, req)
	return args.Get(0).(additemtocart.CartResponse), args.Error(1)
}

func post(uc additemtocart.Usecase, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/carts/me/items", func(c *gin.Context) { c.Set("user_id", buyer); additemtocart.NewHandler(uc).AddItem(c) })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/carts/me/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestAddItemHandler_OK(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("AddItem", mock.Anything, buyer, mock.MatchedBy(func(r additemtocart.AddItemRequest) bool {
		return *r.ProductID == 5 && *r.Quantity == 2
	})).Return(additemtocart.CartResponse{ID: 1}, nil)

	w := post(uc, `{"product_id":5,"quantity":2}`)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestAddItemHandler_InvalidBody(t *testing.T) {
	for _, body := range []string{`{`, `{"product_id":"five","quantity":1}`} {
		uc := new(mockUsecase)
		w := post(uc, body)
		assert.Equal(t, http.StatusBadRequest, w.Code, body)
		uc.AssertNotCalled(t, "AddItem", mock.Anything, mock.Anything, mock.Anything)
	}
}

func TestAddItemHandler_Conflict(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("AddItem", mock.Anything, buyer, mock.Anything).Return(additemtocart.CartResponse{}, apperrors.NewConflict("not enough stock"))

	w := post(uc, `{"product_id":5,"quantity":9}`)

	assert.Equal(t, http.StatusConflict, w.Code)
}
