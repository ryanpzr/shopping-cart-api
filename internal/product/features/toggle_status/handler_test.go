package togglestatus_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	togglestatus "github.com/ryanpzr/shopping-cart-api/internal/product/features/toggle_status"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) Toggle(ctx context.Context, productID, userID int) (togglestatus.ToggleResponse, error) {
	args := m.Called(ctx, productID, userID)
	return args.Get(0).(togglestatus.ToggleResponse), args.Error(1)
}

func patch(uc togglestatus.Usecase, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/products/:id/status", func(c *gin.Context) {
		c.Set("user_id", 7)
		togglestatus.NewHandler(uc).Toggle(c)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, path, nil))
	return w
}

func TestToggleHandler_OK(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Toggle", mock.Anything, 3, 7).Return(togglestatus.ToggleResponse{ID: 3, Status: "inactive"}, nil)

	w := patch(uc, "/products/3/status")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"id":3,"status":"inactive"}`, w.Body.String())
}

func TestToggleHandler_InvalidID(t *testing.T) {
	uc := new(mockUsecase)
	w := patch(uc, "/products/abc/status")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Toggle", mock.Anything, mock.Anything, mock.Anything)
}

func TestToggleHandler_Forbidden(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("Toggle", mock.Anything, 3, 7).
		Return(togglestatus.ToggleResponse{}, apperrors.NewForbidden("you are not the owner of this product"))

	w := patch(uc, "/products/3/status")

	assert.Equal(t, http.StatusForbidden, w.Code)
}
