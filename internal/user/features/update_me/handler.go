package updateme

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Handler interface {
	UpdateMe(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) Handler {
	return &handler{usecase: usecase}
}

func (h *handler) UpdateMe(c *gin.Context) {
	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid request body"))
		return
	}

	if req.Name == "" && req.Email == "" {
		apperrors.HandleError(c, apperrors.NewBadRequest("at least one field (name or email) must be provided"))
		return
	}

	if req.Email != "" && !emailRegex.MatchString(req.Email) {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid email format"))
		return
	}

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

	resp, err := h.usecase.UpdateMe(userID, req)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
