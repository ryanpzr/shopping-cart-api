package product

import (
	"github.com/gin-gonic/gin"
	createproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/create_product"
	deleteproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/delete_product"
	getproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/get_product"
	listproducts "github.com/ryanpzr/shopping-cart-api/internal/product/features/list_products"
	togglestatus "github.com/ryanpzr/shopping-cart-api/internal/product/features/toggle_status"
	updateproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/update_product"
)

type ProductHandlers struct {
	List   listproducts.Handler
	Get    getproduct.Handler
	Create createproduct.Handler
	Update updateproduct.Handler
	Toggle togglestatus.Handler
	Delete deleteproduct.Handler
}

// MapRouters registra rotas de produto.
// public  → sem autenticação (GET list e GET by id)
// protected → requer middleware.Auth() aplicado externamente
func MapRouters(public, protected *gin.RouterGroup, h ProductHandlers) {
	public.GET("", h.List.List)
	public.GET("/:id", h.Get.Get)

	protected.POST("", h.Create.Create)
	protected.PUT("/:id", h.Update.Update)
	protected.PATCH("/:id/status", h.Toggle.Toggle)
	protected.DELETE("/:id", h.Delete.Delete)
}
