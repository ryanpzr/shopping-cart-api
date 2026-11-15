package deleteproduct_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	deleteproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/delete_product"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Delete(ctx context.Context, productID, userID int, role string) error {
	return m.Called(ctx, productID, userID, role).Error(0)
}

func del(uc deleteproduct.Usecase, path, role string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/products/:id", func(c *gin.Context) {
		c.Set("user_id", 7)
		c.Set("role", role)
		deleteproduct.NewHandler(uc).Delete(c)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, path, nil))
	return w
}

func TestDeleteHandler_PassesUserAndRole(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Delete", mock.Anything, 3, 7, "admin").Return(nil)

	w := del(uc, "/products/3", "admin")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"product deleted successfully"}`, w.Body.String())
	uc.AssertExpectations(t)
}

func TestDeleteHandler_InvalidID(t *testing.T) {
	uc := new(mockUsecase)
	w := del(uc, "/products/abc", "client")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleteHandler_Conflict(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Delete", mock.Anything, 3, 7, "client").
		Return(apperrors.NewConflict("product has active orders and cannot be deleted"))

	w := del(uc, "/products/3", "client")

	assert.Equal(t, http.StatusConflict, w.Code)
}
