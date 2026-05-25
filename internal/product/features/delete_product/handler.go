package deleteproduct

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	Delete(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

func (h *handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid product id"))
		return
	}

	userID, _ := c.Get("user_id")
	userIDInt, _ := userID.(int)

	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	if err := h.usecase.Delete(id, userIDInt, roleStr); err != nil {
		apperrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, DeleteResponse{Message: "product deleted successfully"})
}
