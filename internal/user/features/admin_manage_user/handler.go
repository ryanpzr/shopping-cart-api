package adminmanageuser

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Handler interface {
	BanUser(c *gin.Context)
	TimeoutUser(c *gin.Context)
	UnbanUser(c *gin.Context)
}

type handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) Handler {
	return &handler{usecase: usecase}
}

func parseIDParam(c *gin.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return 0, apperrors.NewBadRequest("invalid user id")
	}
	return id, nil
}

func (h *handler) BanUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	resp, err := h.usecase.BanUser(id)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *handler) TimeoutUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}

	var req TimeoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.HandleError(c, apperrors.NewBadRequest("invalid request body"))
		return
	}
	if req.DurationHours <= 0 {
		apperrors.HandleError(c, apperrors.NewBadRequest("duration_hours must be greater than 0"))
		return
	}

	resp, err := h.usecase.TimeoutUser(id, req.DurationHours)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *handler) UnbanUser(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	resp, err := h.usecase.UnbanUser(id)
	if err != nil {
		apperrors.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
