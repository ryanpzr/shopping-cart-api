package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/controller"
)

func ClientRouters(c *gin.RouterGroup) {
	client := c.Group("/client")
	{
		availableProducts := client.Group("/available/products")
		{
			availableProducts.GET("/all", controller.GetAllProductsAvailable)
		}

		cart := client.Group("/cart")
		{
			cart.POST("/add", controller.PostProductCart)
			cart.DELETE("/delete/:productId", controller.DeleteProductCart)
			cart.GET("/all", controller.GetAllProductsCart)
			cart.PUT("/change/:productId/increase", controller.IncreaseProductCart)
			cart.PUT("/change/:productId/decrease", controller.DecreaseProductCart)
		}
	}
}
