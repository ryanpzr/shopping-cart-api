package product

import (
	"github.com/gin-gonic/gin"
	changeinfoproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/change_info_product"
)

func MapRouters(r *gin.RouterGroup, hd changeinfoproduct.Handler) {
	r.GET("", hd.GetAllProduct)
	r.POST("", hd.PostNewProduct)
	r.PUT("/:productId", hd.ChangeInfoProduct)
}
