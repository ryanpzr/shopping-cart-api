package createproduct_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	createproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/create_product"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Create(ctx context.Context, sellerID int, req createproduct.CreateRequest) (createproduct.ProductResponse, error) {
	args := m.Called(ctx, sellerID, req)
	return args.Get(0).(createproduct.ProductResponse), args.Error(1)
}

func post(uc createproduct.Usecase, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/products", func(c *gin.Context) {
		c.Set("user_id", 7)
		createproduct.NewHandler(uc).Create(c)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestCreateHandler_Created(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Create", mock.Anything, 7, mock.MatchedBy(func(r createproduct.CreateRequest) bool {
		return r.Title == "Lamp" && *r.Price == 30 && *r.Quantity == 2
	})).Return(createproduct.ProductResponse{ID: 1, SellerID: 7, Title: "Lamp", FinalPrice: 30}, nil)

	w := post(uc, `{"title":"Lamp","price":30,"quantity":2}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"final_price":30`)
	uc.AssertExpectations(t)
}

func TestCreateHandler_IgnoresSellerIDFromBody(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Create", mock.Anything, 7, mock.Anything).Return(createproduct.ProductResponse{ID: 1, SellerID: 7}, nil)

	w := post(uc, `{"title":"Lamp","price":30,"quantity":2,"seller_id":99}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	uc.AssertExpectations(t)
}

func TestCreateHandler_InvalidJSON(t *testing.T) {
	uc := new(mockUsecase)

	w := post(uc, `{"title":`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateHandler_WrongFieldType(t *testing.T) {
	w := post(new(mockUsecase), `{"title":"Lamp","price":"thirty","quantity":2}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateHandler_UsecaseError(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Create", mock.Anything, 7, mock.Anything).
		Return(createproduct.ProductResponse{}, apperrors.NewBadRequest("price must be greater than 0"))

	w := post(uc, `{"title":"Lamp","price":0,"quantity":2}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"price must be greater than 0"}`, w.Body.String())
}
