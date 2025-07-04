package service

import (
	"github.com/ryanpzr/shopping-cart-api/internal/model"
	"net/http"
	"strconv"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/repository"
)

type Product interface {
	ChangeInfoProduct(ctx *gin.Context)
	GetAllProduct(ctx *gin.Context)
	PostNewProduct(ctx *gin.Context)
}

type productService struct {
	repo repository.Repository
}

func NewProductService(repo repository.Repository) Product {
	return &productService{repo: repo}
}

func (s *productService) ChangeInfoProduct(ctx *gin.Context) {
	id := ctx.Param("productId")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "product id should be int"})
		return
	}

	var productDTO model.ProductDTO
	err = ctx.BindJSON(&productDTO)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if productDTO.Photo == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product photo can't be nil"})
		return
	}
	if productDTO.Title == nil || *productDTO.Title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product title can't be empty or nil"})
		return
	}
	if productDTO.Description == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product description can't be nil"})
		return
	}
	if productDTO.Price == nil || *productDTO.Price == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product price can't be zero or nil"})
		return
	}
	if productDTO.Quantity == nil || *productDTO.Quantity == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product quantity can't be zero or nil"})
		return
	}

	product := model.Product{
		Photo:       *productDTO.Photo,
		Title:       *productDTO.Title,
		Description: *productDTO.Description,
		Price:       *productDTO.Price,
		Quantity:    *productDTO.Quantity,
	}

	err = s.repo.ChangeInfoProduct(idInt, product)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, "product updated")
}

func (s *productService) GetAllProduct(ctx *gin.Context) {
	products, err := s.repo.GetAllProduct()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get products: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, map[string][]model.Product{
		"data": products,
	})
}

func (s *productService) PostNewProduct(ctx *gin.Context) {
	var productDTO model.ProductDTO
	err := ctx.BindJSON(&productDTO)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if productDTO.Photo == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product photo can't be nil"})
		return
	}
	regex := regexp.MustCompile(`^[A-Za-z]+$`)
	if !regex.MatchString(*productDTO.Title) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Title there isn't numbers"})
		return
	}
	if productDTO.Title == nil || *productDTO.Title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product title can't be empty or nil"})
		return
	}
	if productDTO.Description == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product description can't be nil"})
		return
	}
	if productDTO.Price == nil || *productDTO.Price == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product price can't be zero or nil"})
		return
	}
	if productDTO.Quantity == nil || *productDTO.Quantity == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Product quantity can't be zero or nil"})
		return
	}

	product := model.Product{
		Photo:       *productDTO.Photo,
		Title:       *productDTO.Title,
		Description: *productDTO.Description,
		Price:       *productDTO.Price,
		Quantity:    *productDTO.Quantity,
	}

	result, err := s.repo.PostNewProduct(product)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, map[string]model.Product{
		"data": result,
	})
}
