package additemtocart

import (
	"github.com/gin-gonic/gin"
)

func NewHandler(us Usecase) Handler {
	return &handler{
		us: us,
	}
}

type Handler interface {
	AddItemHandler(ctx *gin.Context)
}

type handler struct {
	us Usecase
}

func (h *handler) AddItemHandler(ctx *gin.Context) {
	//TODO implement me
	panic("implement me")
}
