package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/service"
)

type Client struct {
	cartService service.Cart
}

func NewClient(c service.Cart) *Client {
	return &Client{cartService: c}
}

func (cl *Client) ClientRouters(c *gin.RouterGroup) {
	client := c.Group("/client")
	{
		availableProducts := client.Group("/available/products")
		{
			availableProducts.GET("/all", cl.cartService.GetAllProductsAvailable)
		}

		cart := client.Group("/cart")
		{
			cart.POST("/add", cl.cartService.PostProductCart)
			cart.DELETE("/delete/:productId", cl.cartService.DeleteProductCart)
			cart.GET("/all", cl.cartService.GetAllProductsCart)
			cart.PUT("/change/:productId/increase", cl.cartService.IncreaseProductCart)
			cart.PUT("/change/:productId/decrease", cl.cartService.DecreaseProductCart)
		}
	}
}
