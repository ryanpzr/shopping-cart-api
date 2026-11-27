package changeitemquantity

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	ChangeQuantity(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

func (h *handler) ChangeQuantity(c *gin.Context) {
	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil || itemID < 1 {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid item id"))
		return
	}

	var req ChangeQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid request body"))
		return
	}

	resp, err := h.usecase.ChangeQuantity(c.Request.Context(), c.GetInt("user_id"), itemID, req)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
