package additemtocart

import (
	"github.com/gin-gonic/gin"
	repository "github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

func Setup(r *gin.Engine) Handler {
	rp := repository.NewRepository(nil)
	us := NewUsecase(rp)
	return NewHandler(us)
}
