package setup

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/app/route"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
	changeinfoproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/change_info_product"
	productshared "github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	usershared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"
)

func Setup(api *gin.RouterGroup, conn *sql.DB) {
	rpUser := usershared.NewRepository(conn)
	usRegister := register.NewUsecase(rpUser)
	usLogin := login.NewUsecase(rpUser)
	hdRegister := register.NewHandler(usRegister)
	hdLogin := login.NewHandler(usLogin)

	rpProduct := productshared.NewRepository(conn)
	usProduct := changeinfoproduct.NewUsecase(rpProduct)
	hdProduct := changeinfoproduct.NewHandler(usProduct)

	route.RegisterRoutes(api, hdRegister, hdLogin, hdProduct)
}
