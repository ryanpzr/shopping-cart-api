package setup

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/app/route"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
	changeinfoproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/change_info_product"
	productshared "github.com/ryanpzr/shopping-cart-api/internal/product/shared"
	adminactivitylog "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_activity_log"
	admingetuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_get_user"
	adminlistusers "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_list_users"
	adminmanageuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_manage_user"
	getme "github.com/ryanpzr/shopping-cart-api/internal/user/features/get_me"
	updateme "github.com/ryanpzr/shopping-cart-api/internal/user/features/update_me"
	usershared "github.com/ryanpzr/shopping-cart-api/internal/user/shared"
)

func Setup(api *gin.RouterGroup, conn *sql.DB) {
	// User repository (shared across auth and user features)
	rpUser := usershared.NewRepository(conn)

	// Auth features
	usRegister := register.NewUsecase(rpUser)
	usLogin := login.NewUsecase(rpUser)
	hdRegister := register.NewHandler(usRegister)
	hdLogin := login.NewHandler(usLogin)

	// Product features
	rpProduct := productshared.NewRepository(conn)
	usProduct := changeinfoproduct.NewUsecase(rpProduct)
	hdProduct := changeinfoproduct.NewHandler(usProduct)

	// User features — client
	usGetMe := getme.NewUsecase(rpUser)
	usUpdateMe := updateme.NewUsecase(rpUser)
	hdGetMe := getme.NewHandler(usGetMe)
	hdUpdateMe := updateme.NewHandler(usUpdateMe)

	// User features — admin
	usAdminList := adminlistusers.NewUsecase(rpUser)
	usAdminGet := admingetuser.NewUsecase(rpUser)
	usManage := adminmanageuser.NewUsecase(rpUser)
	usActivity := adminactivitylog.NewUsecase() // stub: no repository until module 06
	hdAdminList := adminlistusers.NewHandler(usAdminList)
	hdAdminGet := admingetuser.NewHandler(usAdminGet)
	hdManage := adminmanageuser.NewHandler(usManage)
	hdActivity := adminactivitylog.NewHandler(usActivity)

	route.RegisterRoutes(
		api,
		hdRegister, hdLogin,
		hdProduct,
		hdGetMe, hdUpdateMe,
		hdAdminList, hdAdminGet, hdManage, hdActivity,
	)
}
