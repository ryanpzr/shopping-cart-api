package apperrors

import (
	"errors"
	"net/http"
)

func validateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrBadRequest):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case errors.Is(err, ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	case errors.Is(err, ErrConflict):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case errors.Is(err, ErrNoContent):
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	case errors.Is(err, ErrTooManyRequests):
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	case errors.Is(err, ErrInternalServer):
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
