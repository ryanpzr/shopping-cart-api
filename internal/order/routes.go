package order

import (
	"github.com/gin-gonic/gin"
	adminupdatestatus "github.com/ryanpzr/shopping-cart-api/internal/order/features/admin_update_status"
	cancelorder "github.com/ryanpzr/shopping-cart-api/internal/order/features/cancel_order"
	getorder "github.com/ryanpzr/shopping-cart-api/internal/order/features/get_order"
	listorders "github.com/ryanpzr/shopping-cart-api/internal/order/features/list_orders"
)

type OrderHandlers struct {
	List        listorders.Handler
	Get         getorder.Handler
	Cancel      cancelorder.Handler
	AdminStatus adminupdatestatus.Handler
}

// MapClientRoutes registers the order routes under /orders. The caller must
// apply middleware.Auth(): orders always belong to the authenticated user
// (an admin may read any order through Get).
func MapClientRoutes(r *gin.RouterGroup, h OrderHandlers) {
	r.GET("/me", h.List.List)
	r.GET("/:id", h.Get.Get)
	r.PATCH("/:id/cancel", h.Cancel.Cancel)
}

// MapAdminRoutes registers the admin order routes under /admin/orders. The
// caller must apply middleware.Auth() and middleware.RequireRole("admin").
func MapAdminRoutes(r *gin.RouterGroup, h OrderHandlers) {
	r.PATCH("/:id/status", h.AdminStatus.UpdateStatus)
}
