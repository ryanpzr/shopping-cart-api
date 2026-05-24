package getproduct

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	Get(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

func (h *handler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid product id"))
		return
	}

	resp, err := h.usecase.Get(id)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
