package register

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	Register(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(uc Usecase) Handler {
	return &handler{usecase: uc}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func (h *handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid request body"))
		return
	}
	if req.Name == "" {
		apperrors.HandleError(c, apperrors.NewBadRequest("name is required"))
		return
	}
	if !emailRegex.MatchString(req.Email) {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid email"))
		return
	}
	if len(req.Password) < 8 {
		apperrors.HandleError(c, apperrors.NewBadRequest("password must be at least 8 characters"))
		return
	}

	resp, err := h.usecase.Register(req)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}
