package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	jwtutil "github.com/ryanpzr/shopping-cart-api/pkg/jwt"
	"github.com/ryanpzr/shopping-cart-api/pkg/middleware"
)

func newProtectedRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", middleware.Auth(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})
	return r
}

func TestAuth_MissingHeader(t *testing.T) {
	r := newProtectedRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	r := newProtectedRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.invalid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ValidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	token, _ := jwtutil.Generate(42, "client", "user@example.com")

	r := newProtectedRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func newRoleRouter(allowedRoles ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", middleware.Auth(), middleware.RequireRole(allowedRoles...), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestRequireRole_Allowed(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	token, _ := jwtutil.Generate(1, "admin", "admin@example.com")

	r := newRoleRouter("admin")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_Forbidden(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	token, _ := jwtutil.Generate(1, "client", "user@example.com")

	r := newRoleRouter("admin")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
