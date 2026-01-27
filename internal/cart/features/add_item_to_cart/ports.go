package additemtocart

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/shared"
)

type Usecase interface {
	Execute(ctx *gin.Context) error
}

type Repository = shared.Repository
