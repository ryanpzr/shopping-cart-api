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

	sellerID := c.GetInt("user_id")

	resp, err := h.usecase.Create(c.Request.Context(), sellerID, req)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}
