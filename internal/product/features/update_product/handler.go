package updateproduct

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	Update(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

func (h *handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid product id"))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid request body"))
		return
	}

	if req.Price != nil && *req.Price <= 0 {
		apperrors.HandleError(c, apperrors.NewBadRequest("price must be greater than 0"))
		return
	}
	if req.Price != nil && *req.Price > 1_000_000 {
		apperrors.HandleError(c, apperrors.NewBadRequest("price must not exceed 1,000,000"))
		return
	}
	if req.DiscountPercentage != nil && (*req.DiscountPercentage < 0 || *req.DiscountPercentage > 80) {
		apperrors.HandleError(c, apperrors.NewBadRequest("discount_percentage must be between 0 and 80"))
		return
	}
	if req.Quantity != nil && *req.Quantity < 1 {
		apperrors.HandleError(c, apperrors.NewBadRequest("quantity must be at least 1"))
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt, _ := userID.(int)

	resp, err := h.usecase.Update(id, userIDInt, req)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
