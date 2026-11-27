package setup

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/internal/app/route"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/login"
	"github.com/ryanpzr/shopping-cart-api/internal/auth/features/register"
	"github.com/ryanpzr/shopping-cart-api/internal/cart"
	additemtocart "github.com/ryanpzr/shopping-cart-api/internal/cart/features/add_item_to_cart"
	changeitemquantity "github.com/ryanpzr/shopping-cart-api/internal/cart/features/change_item_quantity"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/features/checkout"
	getcart "github.com/ryanpzr/shopping-cart-api/internal/cart/features/get_cart"
	removeitem "github.com/ryanpzr/shopping-cart-api/internal/cart/features/remove_item"
	cartshared "github.com/ryanpzr/shopping-cart-api/internal/cart/shared"
	"github.com/ryanpzr/shopping-cart-api/internal/product"
	createproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/create_product"
	deleteproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/delete_product"
	getproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/get_product"
	listproducts "github.com/ryanpzr/shopping-cart-api/internal/product/features/list_products"
	togglestatus "github.com/ryanpzr/shopping-cart-api/internal/product/features/toggle_status"
	updateproduct "github.com/ryanpzr/shopping-cart-api/internal/product/features/update_product"
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
	hdProducts := product.ProductHandlers{
		List:   listproducts.NewHandler(listproducts.NewUsecase(rpProduct)),
		Get:    getproduct.NewHandler(getproduct.NewUsecase(rpProduct)),
		Create: createproduct.NewHandler(createproduct.NewUsecase(rpProduct)),
		Update: updateproduct.NewHandler(updateproduct.NewUsecase(rpProduct)),
		Toggle: togglestatus.NewHandler(togglestatus.NewUsecase(rpProduct)),
		Delete: deleteproduct.NewHandler(deleteproduct.NewUsecase(rpProduct, productshared.NewOrderRepository(conn))),
	}

	// Cart features: mutations and checkout run through a unit of work, so
	// cart, product and order repositories share one transaction.
	rpCart := cartshared.NewRepository(conn)
	uow := cartshared.NewUnitOfWork(conn)
	hdCart := cart.CartHandlers{
		Get:            getcart.NewHandler(getcart.NewUsecase(rpCart)),
		AddItem:        additemtocart.NewHandler(additemtocart.NewUsecase(rpCart, uow)),
		ChangeQuantity: changeitemquantity.NewHandler(changeitemquantity.NewUsecase(uow)),
		RemoveItem:     removeitem.NewHandler(removeitem.NewUsecase(uow)),
		Checkout:       checkout.NewHandler(checkout.NewUsecase(uow)),
	}

	// User features — client
	hdGetMe := getme.NewHandler(getme.NewUsecase(rpUser))
	hdUpdateMe := updateme.NewHandler(updateme.NewUsecase(rpUser))

	// User features — admin
	hdAdminList := adminlistusers.NewHandler(adminlistusers.NewUsecase(rpUser))
	hdAdminGet := admingetuser.NewHandler(admingetuser.NewUsecase(rpUser))
	hdManage := adminmanageuser.NewHandler(adminmanageuser.NewUsecase(rpUser))
	hdActivity := adminactivitylog.NewHandler(adminactivitylog.NewUsecase()) // stub until module 06

	route.RegisterRoutes(
		api,
		hdRegister, hdLogin,
		hdProducts,
		hdCart,
		hdGetMe, hdUpdateMe,
		hdAdminList, hdAdminGet, hdManage, hdActivity,
	)
}
