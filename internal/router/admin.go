package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/service"
)

type Admin struct {
	productService service.Product
}

func NewAdmin(ps service.Product) *Admin {
	return &Admin{productService: ps}
}

func (a *Admin) AdminRouters(rg *gin.RouterGroup) {
	admin := rg.Group("/admin")
	{
		admin.GET("/product/all", a.productService.GetAllProduct)
		admin.POST("/product", a.productService.PostNewProduct)
		admin.PUT("/product/:productId", a.productService.ChangeInfoProduct)
	}
}
