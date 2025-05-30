package service

import (
	"github.com/gin-gonic/gin"
)

type Product interface {
	ChangeInfoProduct(ctx *gin.Context)
	GetAllProduct(ctx *gin.Context)
	PostNewProduct(ctx *gin.Context)
}

type productService struct{}

func NewProductService() Product {
	return &productService{}
}

func (s *productService) ChangeInfoProduct(ctx *gin.Context) {
}

func (s *productService) GetAllProduct(ctx *gin.Context) {
}

func (s *productService) PostNewProduct(ctx *gin.Context) {
}
