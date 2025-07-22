package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/handler"
)

type Admin struct {
	hd handler.HandlerProduct
}

func NewAdmin(hd handler.HandlerProduct) *Admin {
	return &Admin{hd: hd}
}

func (a *Admin) AdminRouters(rg *gin.RouterGroup) {
	admin := rg.Group("/admin")
	{
		admin.GET("/product/all", a.hd.GetAllProduct)
		admin.POST("/product", a.hd.PostNewProduct)
		admin.PUT("/product/:productId", a.hd.ChangeInfoProduct)
	}
}
