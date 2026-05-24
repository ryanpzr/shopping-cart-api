package route

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/auth"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
	"github.com/ryanpzr/shopping-cart-api/internal/product"
	changeinfoproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/change_info_product"
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
	hdProduct changeinfoproduct.Handler,
	hdGetMe getme.Handler,
	hdUpdateMe updateme.Handler,
	hdAdminList adminlistusers.Handler,
	hdAdminGet admingetuser.Handler,
	hdManage adminmanageuser.Handler,
	hdActivity adminactivitylog.Handler,
) {
	// Public auth routes
	authGroup := api.Group("/auth")
	auth.MapRouters(authGroup, hdRegister, hdLogin)

	// Protected routes — any authenticated role
	protected := api.Group("")
	protected.Use(middleware.Auth())
	product.MapRouters(protected.Group("/products"), hdProduct)
	user.MapClientRoutes(protected.Group("/users"), hdGetMe, hdUpdateMe)

	// Admin-only routes
	adminProtected := api.Group("")
	adminProtected.Use(middleware.Auth(), middleware.RequireRole("admin"))
	user.MapAdminRoutes(adminProtected.Group("/admin/users"), hdAdminList, hdAdminGet, hdManage, hdActivity)
}
