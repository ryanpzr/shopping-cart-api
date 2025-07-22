package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/service"
)

func NewCartHandler(sv service.Cart) HandlerCart {
	return &handlerCart{
		sv: sv,
	}
}

type HandlerCart interface {
	GetAllProductsAvailable(ctx *gin.Context)
	PostProductCart(ctx *gin.Context)
	DeleteProductCart(ctx *gin.Context)
	GetAllProductsCart(ctx *gin.Context)
	IncreaseProductCart(ctx *gin.Context)
	DecreaseProductCart(ctx *gin.Context)
}

type handlerCart struct {
	sv service.Cart
}

func (h handlerCart) GetAllProductsAvailable(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h handlerCart) PostProductCart(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h handlerCart) DeleteProductCart(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h handlerCart) GetAllProductsCart(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h handlerCart) IncreaseProductCart(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h handlerCart) DecreaseProductCart(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}
