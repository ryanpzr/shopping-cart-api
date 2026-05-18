package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
)

func MapRouters(r *gin.RouterGroup, hdRegister register.Handler, hdLogin login.Handler) {
	r.POST("/register", hdRegister.Register)
	r.POST("/login", hdLogin.Login)
}
