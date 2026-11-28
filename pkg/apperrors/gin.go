package apperrors

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const internalServerMessage = "internal server error"

// HandleError writes err as a JSON error response.
// Client errors (4xx) expose their message. Server errors (5xx) and unknown
// errors are logged and answered with a generic message, so database or
// driver details never reach the client. The one exception is 503: its message
// is a fixed text written by the usecase ("X is temporarily unavailable"),
// never a driver error, so it is safe to show and tells the client to retry.
func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Code < http.StatusInternalServerError {
		c.JSON(appErr.Code, gin.H{"error": appErr.Message})
		return
	}

	code, message := http.StatusInternalServerError, internalServerMessage
	if appErr != nil {
		code = appErr.Code
		if code == http.StatusServiceUnavailable {
			message = appErr.Message
		}
	}
	log.Printf("%s %s: %v", c.Request.Method, c.FullPath(), err)
	c.JSON(code, gin.H{"error": message})
}
