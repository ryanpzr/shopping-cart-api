package adminupdatestatus

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	UpdateStatus(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

func (h *handler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid order id"))
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid request body"))
		return
	}

	resp, err := h.usecase.UpdateStatus(c.Request.Context(), id, req)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
