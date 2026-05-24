package adminlistusers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	ListUsers(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) Handler {
	return &handler{usecase: usecase}
}

func (h *handler) ListUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	resp, err := h.usecase.ListUsers(page, limit)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
