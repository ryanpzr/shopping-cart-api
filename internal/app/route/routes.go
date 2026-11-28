package route

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/auth"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
	"github.com/ryanpzr/shopping-cart-api/internal/cart"
	"github.com/ryanpzr/shopping-cart-api/internal/order"
	"github.com/ryanpzr/shopping-cart-api/internal/product"
	"github.com/ryanpzr/shopping-cart-api/internal/user"
	adminactivitylog "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_activity_log"
	admingetuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_get_user"
	adminlistusers "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_list_users"
	adminmanageuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_manage_user"
	getme "github.com/ryanpzr/shopping-cart-api/internal/user/features/get_me"
	updateme "github.com/ryanpzr/shopping-cart-api/internal/user/features/update_me"
	"github.com/ryanpzr/shopping-cart-api/pkg/middleware"
)

func RegisterRoutes(
	api *gin.RouterGroup,
	hdRegister register.Handler,
	hdLogin login.Handler,
	hdProducts product.ProductHandlers,
	hdCart cart.CartHandlers,
	hdOrders order.OrderHandlers,
	hdGetMe getme.Handler,
	hdUpdateMe updateme.Handler,
	hdAdminList adminlistusers.Handler,
	hdAdminGet admingetuser.Handler,
	hdManage adminmanageuser.Handler,
	hdActivity adminactivitylog.Handler,
) {
	// Public: auth
	authGroup := api.Group("/auth")
	auth.MapRouters(authGroup, hdRegister, hdLogin)

	// Products: GET routes are public, the rest require auth
	productPublic := api.Group("/products")
	productProtected := api.Group("/products")
	productProtected.Use(middleware.Auth())
	product.MapRouters(productPublic, productProtected, hdProducts)

	// Protected: any authenticated role
	protected := api.Group("")
	protected.Use(middleware.Auth())
	user.MapClientRoutes(protected.Group("/users"), hdGetMe, hdUpdateMe)
	cart.MapRouters(protected.Group("/carts/me"), hdCart)
	order.MapClientRoutes(protected.Group("/orders"), hdOrders)

	// Admin only
	adminProtected := api.Group("")
	adminProtected.Use(middleware.Auth(), middleware.RequireRole("admin"))
	user.MapAdminRoutes(adminProtected.Group("/admin/users"), hdAdminList, hdAdminGet, hdManage, hdActivity)
	order.MapAdminRoutes(adminProtected.Group("/admin/orders"), hdOrders)
}
