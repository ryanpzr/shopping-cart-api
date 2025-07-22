package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/handler"
)

type Client struct {
	hd handler.HandlerCart
}

func NewClient(hd handler.HandlerCart) *Client {
	return &Client{hd: hd}
}

func (cl *Client) ClientRouters(c *gin.RouterGroup) {
	client := c.Group("/client")
	{
		availableProducts := client.Group("/available/products")
		{
			availableProducts.GET("/all", cl.hd.GetAllProductsAvailable)
		}

		cart := client.Group("/cart")
		{
			cart.POST("/add", cl.hd.PostProductCart)
			cart.DELETE("/delete/:productId", cl.hd.DeleteProductCart)
			cart.GET("/all", cl.hd.GetAllProductsCart)
			cart.PUT("/change/:productId/increase", cl.hd.IncreaseProductCart)
			cart.PUT("/change/:productId/decrease", cl.hd.DecreaseProductCart)
		}
	}
}
