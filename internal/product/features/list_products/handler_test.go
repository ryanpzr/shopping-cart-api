package listproducts_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	listproducts "github.com/ryanpzr/shopping-cart-api/internal/product/features/list_products"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) List(ctx context.Context, req listproducts.ListRequest) (listproducts.ListResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(listproducts.ListResponse), args.Error(1)
}

func list(uc listproducts.Usecase, url string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/products", listproducts.NewHandler(uc).List)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
	return w
}

func TestListHandler_BindsQueryParams(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("List", mock.Anything, mock.MatchedBy(func(r listproducts.ListRequest) bool {
		return *r.Category == "books" && *r.MinPrice == 10 && *r.MaxPrice == 99.9 &&
			*r.Search == "go" && r.Limit == 5 && r.Offset == 10
	})).Return(listproducts.ListResponse{Data: []listproducts.ProductResponse{}}, nil)

	w := list(uc, "/products?category=books&min_price=10&max_price=99.9&search=go&limit=5&offset=10")

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestListHandler_DefaultsWithoutParams(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("List", mock.Anything, listproducts.ListRequest{Limit: 20}).
		Return(listproducts.ListResponse{Data: []listproducts.ProductResponse{}}, nil)

	w := list(uc, "/products")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":[],"meta":{"total":0,"limit":0,"offset":0}}`, w.Body.String())
	uc.AssertExpectations(t)
}

func TestListHandler_InvalidQueryType(t *testing.T) {
	for _, url := range []string{"/products?min_price=abc", "/products?limit=ten"} {
		uc := new(mockUsecase)
		w := list(uc, url)
		assert.Equal(t, http.StatusBadRequest, w.Code, url)
		uc.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
	}
}
