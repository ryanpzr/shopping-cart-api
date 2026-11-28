package listorders_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	listorders "github.com/ryanpzr/shopping-cart-api/internal/order/features/list_orders"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) List(ctx context.Context, buyerID int, req listorders.ListRequest) (listorders.ListResponse, error) {
	args := m.Called(ctx, buyerID, req)
	return args.Get(0).(listorders.ListResponse), args.Error(1)
}

func list(uc listorders.Usecase, url string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/orders/me", func(c *gin.Context) { c.Set("user_id", 7); listorders.NewHandler(uc).List(c) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
	return w
}

func TestListHandler_UsesAuthenticatedUserAndQuery(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("List", mock.Anything, 7, listorders.ListRequest{Limit: 5, Offset: 10}).
		Return(listorders.ListResponse{Data: []listorders.OrderResponse{}, Meta: listorders.MetaResponse{Total: 0, Limit: 5, Offset: 10}}, nil)

	w := list(uc, "/orders/me?limit=5&offset=10")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":[],"meta":{"total":0,"limit":5,"offset":10}}`, w.Body.String())
}

func TestListHandler_DefaultsWithoutParams(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("List", mock.Anything, 7, listorders.ListRequest{Limit: 20}).
		Return(listorders.ListResponse{Data: []listorders.OrderResponse{}}, nil)

	assert.Equal(t, http.StatusOK, list(uc, "/orders/me").Code)
	uc.AssertExpectations(t)
}

func TestListHandler_InvalidQueryType(t *testing.T) {
	for _, url := range []string{"/orders/me?limit=ten", "/orders/me?offset=x"} {
		uc := new(mockUsecase)
		assert.Equal(t, http.StatusBadRequest, list(uc, url).Code, url)
		uc.AssertNotCalled(t, "List", mock.Anything, mock.Anything, mock.Anything)
	}
}
