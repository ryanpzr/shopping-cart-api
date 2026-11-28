package apperrors_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

func serve(err error) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) { apperrors.HandleError(c, err) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	return w
}

func TestHandleError_ClientErrorExposesMessage(t *testing.T) {
	w := serve(apperrors.NewNotFound("product not found"))
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"product not found"}`, w.Body.String())
}

func TestHandleError_ServerErrorHidesDetails(t *testing.T) {
	w := serve(apperrors.NewInternalServer(`pq: relation "products" does not exist`))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"internal server error"}`, w.Body.String())
}

func TestHandleError_UnknownErrorIsInternal(t *testing.T) {
	w := serve(errors.New("boom"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"internal server error"}`, w.Body.String())
}

func TestHandleError_WrappedAppError(t *testing.T) {
	w := serve(errors.Join(errors.New("context"), apperrors.NewConflict("email already registered")))
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.JSONEq(t, `{"error":"email already registered"}`, w.Body.String())
}

func TestHandleError_ServiceUnavailableShowsItsFixedMessage(t *testing.T) {
	w := serve(apperrors.NewServiceUnavailable("activity log is temporarily unavailable"))
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"error":"activity log is temporarily unavailable"}`, w.Body.String())
}

func TestHandleError_OtherServerErrorsStayGeneric(t *testing.T) {
	w := serve(&apperrors.AppError{Code: http.StatusBadGateway, Message: "upstream said: password=secret"})
	assert.Equal(t, http.StatusBadGateway, w.Code)
	assert.JSONEq(t, `{"error":"internal server error"}`, w.Body.String())
}
