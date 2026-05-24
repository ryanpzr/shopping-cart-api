package user

import (
	"github.com/gin-gonic/gin"
	adminactivitylog "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_activity_log"
	admingetuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_get_user"
	adminlistusers "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_list_users"
	adminmanageuser "github.com/ryanpzr/shopping-cart-api/internal/user/features/admin_manage_user"
	getme "github.com/ryanpzr/shopping-cart-api/internal/user/features/get_me"
	updateme "github.com/ryanpzr/shopping-cart-api/internal/user/features/update_me"
)

func MapClientRoutes(
	r *gin.RouterGroup,
	hdGetMe getme.Handler,
	hdUpdateMe updateme.Handler,
) {
	r.GET("/me", hdGetMe.GetMe)
	r.PUT("/me", hdUpdateMe.UpdateMe)
}

func MapAdminRoutes(
	r *gin.RouterGroup,
	hdList adminlistusers.Handler,
	hdGet admingetuser.Handler,
	hdManage adminmanageuser.Handler,
	hdActivity adminactivitylog.Handler,
) {
	r.GET("", hdList.ListUsers)
	r.GET("/:id", hdGet.GetUser)
	r.PATCH("/:id/ban", hdManage.BanUser)
	r.PATCH("/:id/timeout", hdManage.TimeoutUser)
	r.PATCH("/:id/unban", hdManage.UnbanUser)
	r.GET("/:id/activity", hdActivity.GetActivityLog)
}
