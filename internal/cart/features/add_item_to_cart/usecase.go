package additemtocart

import (
	"github.com/gin-gonic/gin"
)

type usecase struct {
	rp Repository
}

func NewUsecase(rp Repository) Usecase {
	return &usecase{
		rp: rp,
	}
}

func (u *usecase) Execute(ctx *gin.Context) error {
	//TODO implement me
	panic("implement me")
}
