package route

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/auth"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
	"github.com/ryanpzr/shopping-cart-api/internal/product"
	changeinfoproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/change_info_product"
	"github.com/ryanpzr/shopping-cart-api/pkg/middleware"
)

func RegisterRoutes(
	api *gin.RouterGroup,
	hdRegister register.Handler,
	hdLogin login.Handler,
	hdProduct changeinfoproduct.Handler,
) {
	authGroup := api.Group("/auth")
	auth.MapRouters(authGroup, hdRegister, hdLogin)

	protected := api.Group("")
	protected.Use(middleware.Auth())
	product.MapRouters(protected.Group("/products"), hdProduct)
}
