package cancelorder

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	Cancel(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

func (h *handler) Cancel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid order id"))
		return
	}

	resp, err := h.usecase.Cancel(c.Request.Context(), id, c.GetInt("user_id"))
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
