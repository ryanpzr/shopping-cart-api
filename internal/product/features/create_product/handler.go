package createproduct

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	Create(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

func (h *handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid request body"))
		return
	}

	if req.Title == "" {
		apperrors.HandleError(c, apperrors.NewBadRequest("title is required"))
		return
	}
	if req.Price <= 0 {
		apperrors.HandleError(c, apperrors.NewBadRequest("price must be greater than 0"))
		return
	}
	if req.Quantity < 0 {
		apperrors.HandleError(c, apperrors.NewBadRequest("quantity must be >= 0"))
		return
	}
	if req.DiscountPercentage != nil && (*req.DiscountPercentage < 0 || *req.DiscountPercentage > 100) {
		apperrors.HandleError(c, apperrors.NewBadRequest("discount_percentage must be between 0 and 100"))
		return
	}

	sellerID, _ := c.Get("user_id")
	sellerIDInt, _ := sellerID.(int)

	resp, err := h.usecase.Create(sellerIDInt, req)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}
