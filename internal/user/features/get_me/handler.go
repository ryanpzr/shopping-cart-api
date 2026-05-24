package getme

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	GetMe(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) Handler {
	return &handler{usecase: usecase}
}

func (h *handler) GetMe(c *gin.Context) {
	rawID, exists := c.Get("user_id")
	if !exists {
		apperrors.HandleError(c, apperrors.NewUnauthorized("missing token claims"))
		return
	}
	userID, ok := rawID.(int)
	if !ok {
		apperrors.HandleError(c, apperrors.NewUnauthorized("invalid token claims"))
		return
	}

	resp, err := h.usecase.GetMe(userID)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
