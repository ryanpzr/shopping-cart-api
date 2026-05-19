package apperrors

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, gin.H{"error": appErr.Message})
		return
	}
	c.JSON(500, gin.H{"error": "internal server error"})
}
