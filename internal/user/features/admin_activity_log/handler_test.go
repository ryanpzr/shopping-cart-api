package adminactivitylog_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adminactivitylog "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_activity_log"
)

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	uc := adminactivitylog.NewUsecase()
	h := adminactivitylog.NewHandler(uc)
	r.GET("/admin/users/:id/activity", h.GetActivityLog)
	return r
}

func TestActivityLogHandler_ReturnsStub(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users/1/activity", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body adminactivitylog.ActivityLogResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 0, body.Total)
	assert.Equal(t, 0, body.TotalPages)
	assert.NotNil(t, body.Data)
	assert.Empty(t, body.Data)
}

func TestActivityLogHandler_InvalidID(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users/abc/activity", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestActivityLogHandler_PaginationForwarded(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin/users/1/activity?page=3&limit=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body adminactivitylog.ActivityLogResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 3, body.Page)
	assert.Equal(t, 10, body.Limit)
}
