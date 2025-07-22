package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/model"
	"github.com/ryanpzr/shopping-cart-api/internal/service"
	"net/http"
	"regexp"
	"strconv"
)

type HandlerProduct interface {
	GetAllProduct(ctx *gin.Context)
	ChangeInfoProduct(ctx *gin.Context)
	PostNewProduct(ctx *gin.Context)
}

type handlerProduct struct {
	productService service.Product
}

func NewHandlerProduct(sv service.Product) HandlerProduct {
	return &handlerProduct{
		productService: sv,
	}
}

func (h *handlerProduct) GetAllProduct(ctx *gin.Context) {
	products, err := h.productService.GetAllProduct()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get products: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, map[string][]model.Product{
		"data": products,
	})
}

func (h *handlerProduct) PostNewProduct(ctx *gin.Context) {
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

	result, err := h.productService.PostNewProduct(product)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, map[string]model.Product{
		"data": result,
	})
}

func (h *handlerProduct) ChangeInfoProduct(ctx *gin.Context) {
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

	err = h.productService.ChangeInfoProduct(idInt, product)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, "product updated")
}
