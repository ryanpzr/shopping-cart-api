package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/controller"
)

func AdminRouters(a *gin.RouterGroup) {
	admin := a.Group("/admin")
	{
		admin.GET("/product/all", controller.GetAllProduct)
		admin.POST("/product", controller.PostNewProduct)
		admin.PUT("/product/:productId", controller.ChangeInfoProduct)
	}
}
