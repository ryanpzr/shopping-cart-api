package setup

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/app/route"
	changeinfoproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/change_info_product"
	productshared "github.com/ryanpzr/shopping-cart-api/internal/product/shared"
)

func Setup(api *gin.RouterGroup, conn *sql.DB) {
	rpProduct := productshared.NewRepository(conn)
	usProduct := changeinfoproduct.NewUsecase(rpProduct)
	hdProduct := changeinfoproduct.NewHandler(usProduct)
	route.RegisterRoutes(api, hdProduct)
}
