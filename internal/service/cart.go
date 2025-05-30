package service

import "github.com/gin-gonic/gin"

type Cart interface {
	GetAllProductsAvailable(ctx *gin.Context)
	GetAllProductsCart(ctx *gin.Context)
	PostProductCart(ctx *gin.Context)
	DeleteProductCart(ctx *gin.Context)
	IncreaseProductCart(ctx *gin.Context)
	DecreaseProductCart(ctx *gin.Context)
}

type cartService struct{}

func NewCartService() Cart {
	return &cartService{}
}

func (s *cartService) GetAllProductsAvailable(ctx *gin.Context) {
}

func (c *cartService) GetAllProductsCart(ctx *gin.Context) {

}

func (c *cartService) PostProductCart(ctx *gin.Context) {

}

func (c *cartService) DeleteProductCart(ctx *gin.Context) {

}

func (c *cartService) IncreaseProductCart(ctx *gin.Context) {

}

func (c *cartService) DecreaseProductCart(ctx *gin.Context) {

}
