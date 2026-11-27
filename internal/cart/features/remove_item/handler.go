package removeitem

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	RemoveItem(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

func (h *handler) RemoveItem(c *gin.Context) {
	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil || itemID < 1 {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid item id"))
		return
	}

	resp, err := h.usecase.RemoveItem(c.Request.Context(), c.GetInt("user_id"), itemID)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
