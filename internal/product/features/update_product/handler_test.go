package updateproduct_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	updateproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/update_product"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Update(ctx context.Context, productID, userID int, req updateproduct.UpdateRequest) (updateproduct.ProductResponse, error) {
	args := m.Called(ctx, productID, userID, req)
	return args.Get(0).(updateproduct.ProductResponse), args.Error(1)
}

func put(uc updateproduct.Usecase, path, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PUT("/products/:id", func(c *gin.Context) {
		c.Set("user_id", 7)
		updateproduct.NewHandler(uc).Update(c)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestUpdateHandler_OK(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Update", mock.Anything, 3, 7, mock.MatchedBy(func(r updateproduct.UpdateRequest) bool {
		return r.Title.Set && *r.Title.Value == "New" && !r.Price.Set
	})).Return(updateproduct.ProductResponse{ID: 3, Title: "New"}, nil)

	w := put(uc, "/products/3", `{"title":"New"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestUpdateHandler_DistinguishesNullFromOmitted(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Update", mock.Anything, 3, 7, mock.MatchedBy(func(r updateproduct.UpdateRequest) bool {
		return r.Category.Set && r.Category.Value == nil && // explicit null
			!r.Photo.Set && // omitted
			r.Price.Set && *r.Price.Value == 9.5
	})).Return(updateproduct.ProductResponse{ID: 3}, nil)

	w := put(uc, "/products/3", `{"category":null,"price":9.5}`)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestUpdateHandler_WrongFieldType(t *testing.T) {
	uc := new(mockUsecase)
	w := put(uc, "/products/3", `{"price":"cheap"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdateHandler_InvalidID(t *testing.T) {
	uc := new(mockUsecase)
	w := put(uc, "/products/x", `{"title":"New"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdateHandler_InvalidBody(t *testing.T) {
	uc := new(mockUsecase)
	w := put(uc, "/products/3", `not json`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdateHandler_Forbidden(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Update", mock.Anything, 3, 7, mock.Anything).
		Return(updateproduct.ProductResponse{}, apperrors.NewForbidden("you are not the owner of this product"))

	w := put(uc, "/products/3", `{"title":"New"}`)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
