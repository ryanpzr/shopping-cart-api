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
	if err != nil || id < 1 {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid product id"))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid request body"))
		return
	}

	resp, err := h.usecase.Update(c.Request.Context(), id, c.GetInt("user_id"), req)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
