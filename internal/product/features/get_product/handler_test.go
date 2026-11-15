package getproduct_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	getproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/get_product"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Get(ctx context.Context, id int) (getproduct.ProductResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(getproduct.ProductResponse), args.Error(1)
}

func get(uc getproduct.Usecase, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/products/:id", getproduct.NewHandler(uc).Get)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestGetHandler_OK(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Get", mock.Anything, 3).Return(getproduct.ProductResponse{ID: 3, Title: "Desk"}, nil)

	w := get(uc, "/products/3")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"title":"Desk"`)
}

func TestGetHandler_InvalidID(t *testing.T) {
	for _, path := range []string{"/products/abc", "/products/0", "/products/-1"} {
		uc := new(mockUsecase)
		w := get(uc, path)
		assert.Equal(t, http.StatusBadRequest, w.Code, path)
		uc.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	}
}

func TestGetHandler_NotFound(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Get", mock.Anything, 9).Return(getproduct.ProductResponse{}, apperrors.NewNotFound("product not found"))

	w := get(uc, "/products/9")

	assert.Equal(t, http.StatusNotFound, w.Code)
}
