package route

import (
	"github.com/gin-gonic/gin"
	changeinfoproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/change_info_product"
	"github.com/ryanpzr/shopping-cart-api/internal/product"
)

func RegisterRoutes(api *gin.RouterGroup, hdProduct changeinfoproduct.Handler) {
	productGroup := api.Group("/products")
	product.MapRouters(productGroup, hdProduct)
}
