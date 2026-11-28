package adminupdatestatus_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	adminupdatestatus "github.com/ryanpzr/shopping-cart-api/internal/order/features/admin_update_status"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type mockUsecase struct{ mock.Mock }

func (m *mockUsecase) UpdateStatus(ctx context.Context, id int, req adminupdatestatus.UpdateStatusRequest) (adminupdatestatus.OrderResponse, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(adminupdatestatus.OrderResponse), args.Error(1)
}

func patch(uc adminupdatestatus.Usecase, path, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.PATCH("/admin/orders/:id/status", adminupdatestatus.NewHandler(uc).UpdateStatus)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestUpdateStatusHandler_OK(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("UpdateStatus", mock.Anything, orderID, mock.MatchedBy(func(r adminupdatestatus.UpdateStatusRequest) bool {
		return *r.Status == "paid"
	})).Return(adminupdatestatus.OrderResponse{ID: orderID, Status: "paid"}, nil)

	w := patch(uc, "/admin/orders/55/status", `{"status":"paid"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"paid"`)
}

func TestUpdateStatusHandler_BadInput(t *testing.T) {
	cases := map[string]string{
		"/admin/orders/abc/status": `{"status":"paid"}`,
		"/admin/orders/0/status":   `{"status":"paid"}`,
		"/admin/orders/55/status":  `{`,
	}
	// A status that is not a string never reaches the usecase.
	cases["/admin/orders/56/status"] = `{"status":5}`
	for path, body := range cases {
		uc := new(mockUsecase)
		assert.Equal(t, http.StatusBadRequest, patch(uc, path, body).Code, path+" "+body)
		uc.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
	}
}

func TestUpdateStatusHandler_ErrorCodes(t *testing.T) {
	uc := new(mockUsecase)
	uc.On("UpdateStatus", mock.Anything, 1, mock.Anything).Return(adminupdatestatus.OrderResponse{}, apperrors.NewConflict("cannot change"))
	uc.On("UpdateStatus", mock.Anything, 2, mock.Anything).Return(adminupdatestatus.OrderResponse{}, apperrors.NewNotFound("order not found"))
	uc.On("UpdateStatus", mock.Anything, 3, mock.Anything).Return(adminupdatestatus.OrderResponse{}, apperrors.NewBadRequest("status is required"))

	assert.Equal(t, http.StatusConflict, patch(uc, "/admin/orders/1/status", `{"status":"paid"}`).Code)
	assert.Equal(t, http.StatusNotFound, patch(uc, "/admin/orders/2/status", `{"status":"paid"}`).Code)
	assert.Equal(t, http.StatusBadRequest, patch(uc, "/admin/orders/3/status", `{}`).Code)
}
